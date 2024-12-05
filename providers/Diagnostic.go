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
)

type DiagnosticData struct {
	Type    uint8  `json:"type"`
	Surname string `json:"surname"`
	Name    string `json:"name"`
}

func PublishDiagnostics(ctx *Ctx, uri Uri, doc *TextDocument) {
	if !supportDiagnostics {
		return
	}

	var err error

	if doc == nil {
		doc, err = TempDoc(uri)

		if err != nil {
			LogDebug("Diagnostic open doc error: %s", err.Error())
			return
		}
	}

	list := make([]proto.Diagnostic, 0)

	for node := range GetErrorNodesIter(doc.Tree.RootNode()) {
		r, err := doc.NodeToRange(node)

		if err != nil {
			LogDebug("Diagnostic error: %s", err.Error())
			return
		}

		list = append(list, proto.Diagnostic{
			Severity: P(proto.DiagnosticSeverityError),
			Range:    *r,
			Message:  L("syntax_error"),
		})
	}

	for _, ref := range root.UnknownRefs {
		if ref.Uri != uri {
			continue
		}

		node := ref.Node
		message := L("unknown_person")
		t := UnknownPersonError

		if IsNameRef(node) {
			f := root.FindFamily(ref.Surname)
			nameNode, surnameNode := GetNameSurname(node)

			if f == nil {
				node = surnameNode
				message = L("unknown_family")
				t = UnknownFamilyError
			} else {
				node = nameNode
				message = L("unknown_person_in_family", f.Name, ToString(nameNode, doc))
			}
		} else if IsNewSurname(node.Parent()) {
			message = L("unknown_family")
			t = UnknownFamilyError
		}

		r, err := doc.NodeToRange(node)

		if err != nil {
			LogDebug("Diagnostic error: %s", err.Error())
			continue
		}

		list = append(list, proto.Diagnostic{
			Severity: P(proto.DiagnosticSeverityError),
			Range:    *r,
			Message:  message,
			Data: DiagnosticData{
				Type: t,
			},
		})
	}

	tempDocs := make(Docs)
	tempDocs[uri] = doc

	for family := range root.FamilyIter() {
		for name, dups := range family.Duplicates {
			member := family.Members[name]
			dups = append(dups, &Duplicate{Member: member})
			count := len(dups)

			var locations []proto.DiagnosticRelatedInformation

			for _, ref := range member.Refs {
				if ref.Uri != uri {
					continue
				}

				r, err := doc.NodeToRange(ToNameNode(ref.Node))

				if err != nil {
					LogDebug("Diagnostic error: %s", err.Error())
					continue
				}

				if locations == nil {
					d, err := tempDocs.Get(family.Uri)

					if err != nil {
						LogDebug("Diagnostic error: %s", err.Error())
						continue
					}

					locations = make([]proto.DiagnosticRelatedInformation, count)
					for i, dup := range dups {
						node := dup.Member.Node
						rr, err := d.NodeToRange(node)

						if err != nil {
							LogDebug("Diagnostic error: %s", err.Error())
							continue
						}

						sources := GetClosestSources(node)

						locations[i] = proto.DiagnosticRelatedInformation{
							Location: proto.Location{
								URI:   family.Uri,
								Range: *rr,
							},
							Message: L("child_of_source", ToString(sources, d)),
						}
					}
				}

				list = append(list, proto.Diagnostic{
					Severity:           P(proto.DiagnosticSeverityWarning),
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

	ctx.Notify(proto.ServerTextDocumentPublishDiagnostics, proto.PublishDiagnosticsParams{
		URI: uri,
		// TODO add version
		Diagnostics: list,
	})
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

		PublishDiagnostics(docDiagnostic.Ctx, uri, doc)
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
