package providers

import (
	"time"

	"github.com/bep/debounce"
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

type DocDebouncer struct {
	Ctx      *Ctx
	Docs     map[Uri]*TextDocument
	Debounce func(func())
}

const (
	UnknownFamilyError = uint8(iota)
	UnknownPersonError
	NameDuplicateWarning
	ChildWithoutRelationsInfo
)

type DiagnosticData struct {
	Type    uint8  `json:"type"`
	Surname string `json:"surname"`
	Name    string `json:"name"`
}

func PublishDiagnostics(ctx *Ctx, uri Uri, doc *TextDocument) error {
	if !supportDiagnostics {
		return nil
	}

	var err error

	if doc == nil {
		doc, err = TempDoc(uri)

		if err != nil {
			return err
		}
	}

	list := make([]proto.Diagnostic, 0)

	add := func(item proto.Diagnostic) {
		list = append(list, item)
	}

	Error := P(proto.DiagnosticSeverityError)
	Warning := P(proto.DiagnosticSeverityWarning)
	Info := P(proto.DiagnosticSeverityInformation)

	for node := range GetErrorNodesIter(doc.Tree.RootNode()) {
		r, err := doc.NodeToRange(node)

		if err != nil {
			return err
		}

		add(proto.Diagnostic{
			Severity: Error,
			Range:    *r,
			Message:  L("syntax_error"),
		})
	}

	for _, ref := range root.UnknownRefs {
		if ref.Uri != uri || ref.Member != nil {
			continue
		}

		node := ref.Node
		message := L("unknown_person", ref.Name)
		t := UnknownPersonError

		if IsNameRef(node) {
			f := root.FindFamily(ref.Surname)
			nameNode, surnameNode := GetNameSurname(node)

			if f == nil {
				node = surnameNode
				message = L("unknown_family", ref.Surname)
				t = UnknownFamilyError
			} else {
				node = nameNode
				message = L("unknown_person_in_family", f.Name, ToString(nameNode, doc))
			}
		} else if IsNewSurname(node) {
			message = L("unknown_family", ref.Surname)
			t = UnknownFamilyError
		}

		r, err := doc.NodeToRange(node)

		if err != nil {
			return err
		}

		add(proto.Diagnostic{
			Severity: Error,
			Range:    *r,
			Message:  message,
			Data: DiagnosticData{
				Type: t,
			},
		})
	}

	tempDocs := make(Docs)
	tempDocs[uri] = doc

	var locations []proto.DiagnosticRelatedInformation

	ensureLocations := func(family *Family, name string) error {
		if locations != nil {
			return nil
		}

		doc, err := tempDocs.Get(family.Uri)

		if err != nil {
			return err
		}

		dups := family.Duplicates[name]

		member := family.GetMember(name)

		dups = append(dups, &Duplicate{Member: member})

		locations = make([]proto.DiagnosticRelatedInformation, len(dups))

		for i, dup := range dups {
			node := dup.Member.Node
			r, err := doc.NodeToRange(node)

			if err != nil {
				return err
			}

			sources := GetClosestSources(node)

			locations[i] = proto.DiagnosticRelatedInformation{
				Location: proto.Location{
					URI:   family.Uri,
					Range: *r,
				},
				Message: L("child_of_source", ToString(sources, doc)),
			}
		}

		return nil
	}

	for family := range root.FamilyIter() {
		for name, dups := range family.Duplicates {
			locations = nil

			member := family.GetMember(name)
			count := len(dups) + 1

			for _, ref := range member.Refs {
				if ref.Uri != uri || ref.Family == family {
					continue
				}

				r, err := doc.NodeToRange(ToNameNode(ref.Node))

				if err != nil {
					return err
				}

				err = ensureLocations(family, name)

				if err != nil {
					return err
				}

				add(proto.Diagnostic{
					Severity:           Warning,
					Range:              *r,
					Message:            L("duplicate_count_of_name", count, name),
					RelatedInformation: locations,
					Data: DiagnosticData{
						Type:    NameDuplicateWarning,
						Surname: family.Name,
						Name:    name,
					},
				})
			}
		}
	}

	if warnChildrenWithoutRelations {
		for mem := range root.MembersIter() {
			if mem.Family.Uri != uri || len(mem.Refs) > 0 || mem.Origin != nil || !IsNameDef(mem.Node.Parent()) {
				continue
			}

			d, err := tempDocs.Get(mem.Family.Uri)

			if err != nil {
				return err
			}

			r, err := d.NodeToRange(mem.Node)

			if err != nil {
				return err
			}

			add(proto.Diagnostic{
				Severity: Info,
				Range:    *r,
				Message:  L("child_without_relations", mem.Name, mem.Family.Name),
				Data: DiagnosticData{
					Type: ChildWithoutRelationsInfo,
				},
			})
		}
	}

	ctx.Notify(proto.ServerTextDocumentPublishDiagnostics, proto.PublishDiagnosticsParams{
		URI: uri,
		// TODO add version
		Diagnostics: list,
	})

	return nil
}

var docDiagnostic = &DocDebouncer{
	Docs:     make(map[Uri]*TextDocument),
	Debounce: debounce.New(200 * time.Millisecond),
}

func (dd *DocDebouncer) Set(uri Uri, doc *TextDocument) {
	d, ok := dd.Docs[uri]

	if doc != nil || d == nil || !ok {
		dd.Docs[uri] = doc
	}

	dd.Debounce(dd.Flush)
}

func (dd *DocDebouncer) Flush() {
	root.UpdateDirty()

	for uri, doc := range dd.Docs {
		delete(dd.Docs, uri)

		if !IsFamilyUri(uri) || !UriFileExist(uri) {
			continue
		}

		err := PublishDiagnostics(docDiagnostic.Ctx, uri, doc)

		if err != nil {
			LogDebug("Diagnostic error: %s", err.Error())
		}
	}
}

func diagnosticOpenDocs(ctx *Ctx) {
	docDiagnostic.Ctx = ctx

	for uri, doc := range GetOpenDocsIter() {
		docDiagnostic.Set(uri, doc)
	}
}

func diagnosticAllDocs(ctx *Ctx) {
	docDiagnostic.Ctx = ctx

	WalkTrees(func(uri Uri, tree *Tree) {
		docDiagnostic.Set(uri, nil)
	})
}

func scheduleDiagnostic(ctx *Ctx, uri Uri, doc *TextDocument) {
	docDiagnostic.Ctx = ctx
	docDiagnostic.Set(uri, doc)
}
