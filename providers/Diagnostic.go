package providers

import (
	"time"

	fm "github.com/redexp/familymarkup-parser"

	"github.com/bep/debounce"
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

type DocDebouncer struct {
	Ctx      *Ctx
	Uris     UriSet
	Debounce func(func())
}

const (
	UnknownFamilyError = uint8(iota)
	UnknownPersonError
	NameDuplicateWarning
	ChildWithoutRelationsInfo
)

func GetDiagnostics(uri Uri) (list []proto.Diagnostic) {
	list = make([]proto.Diagnostic, 0)

	doc, ok := root.Docs[uri]

	if !ok {
		return
	}

	add := func(item proto.Diagnostic) {
		list = append(list, item)
	}

	Error := P(proto.DiagnosticSeverityError)
	Warning := P(proto.DiagnosticSeverityWarning)
	Info := P(proto.DiagnosticSeverityInformation)

	for _, token := range doc.Tokens {
		if token.Type == fm.TokenInvalid || token.ErrType == fm.ErrUnexpected {
			add(proto.Diagnostic{
				Severity: Error,
				Range:    TokenToRange(token),
				Message:  L("syntax_error"),
			})
		}
	}

	refs := root.UnknownRefs

	for _, ref := range refs {
		if ref.Uri != uri {
			continue
		}

		var t uint8
		var loc fm.Loc
		var message string

		p := ref.Person

		switch ref.Type {
		case RefTypeSurname:
			t = UnknownFamilyError
			loc = ref.Token.Loc()
			message = L("unknown_family", ref.Token.Text)

		case RefTypeName:
			t = UnknownPersonError
			loc = p.Name.Loc()
			message = L("unknown_person", p.Name.Text)

		case RefTypeNameSurname:
			f := root.FindFamily(p.Surname.Text)

			if f == nil {
				continue
			}

			t = UnknownPersonError
			loc = p.Name.Loc()
			message = L("unknown_person_in_family", f.Name, p.Name.Text)

		default:
			continue
		}

		add(proto.Diagnostic{
			Severity: Error,
			Range:    LocToRange(loc),
			Message:  message,
			Data: DiagnosticData{
				Type: t,
			},
		})
	}

	var locations []proto.DiagnosticRelatedInformation

	ensureLocations := func(family *Family, member *Member, dups []*Duplicate) {
		if locations != nil {
			return
		}

		doc := GetDoc(family.Uri)

		locations = make([]proto.DiagnosticRelatedInformation, len(dups))

		for i, dup := range dups {
			sources := dup.Member.Person.Relation.Sources

			text := doc.GetTextByLoc(sources.Loc)

			locations[i] = proto.DiagnosticRelatedInformation{
				Location: proto.Location{
					URI:   family.Uri,
					Range: LocToRange(dup.Member.Person.Name.Loc()),
				},
				Message: L("child_of_source", text),
			}
		}
	}

	// duplicates warning
	for family := range root.FamilyIter() {
		for name, dups := range family.Duplicates {
			locations = nil

			member := family.GetMember(name)

			dups = append(dups, &Duplicate{Member: member})

			for ref, refUri := range member.GetRefsIter() {
				p := ref.Person

				if refUri != uri || ref.Type == RefTypeOrigin || p == member.Person || !IsEqNames(name, p.Name.Text) {
					continue
				}

				ensureLocations(family, member, dups)

				add(proto.Diagnostic{
					Severity:           Warning,
					Range:              TokenToRange(p.Name),
					Message:            L("duplicate_count_of_name", len(locations), name),
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
		for f := range root.FamiliesByUriIter(uri) {
			for mem := range f.MembersIter() {
				if !mem.Person.IsChild || mem.HasRef() {
					continue
				}

				add(proto.Diagnostic{
					Severity: Info,
					Range:    TokenToRange(mem.Person.Name),
					Message:  L("child_without_relations", mem.Name, mem.Family.Name),
					Data: DiagnosticData{
						Type: ChildWithoutRelationsInfo,
					},
				})
			}
		}
	}

	return
}

func PublishDiagnostics(ctx *Ctx, uri Uri) {
	if !supportDiagnostics {
		return
	}

	ctx.Notify(proto.ServerTextDocumentPublishDiagnostics, proto.PublishDiagnosticsParams{
		URI:         EncUri(uri),
		Diagnostics: GetDiagnostics(uri),
	})
}

var docDiagnostic = &DocDebouncer{
	Debounce: debounce.New(300 * time.Millisecond),
	Uris:     make(UriSet),
}

func (dd *DocDebouncer) Set(uri Uri) {
	dd.Uris.Set(uri)
	dd.Schedule()
}

func (dd *DocDebouncer) Schedule() {
	dd.Debounce(dd.Flush)
}

func (dd *DocDebouncer) Flush() {
	_ = root.UpdateDirty()

	for uri := range dd.Uris {
		dd.Uris.Remove(uri)

		if _, ok := root.Docs[uri]; !ok {
			continue
		}

		go PublishDiagnostics(docDiagnostic.Ctx, uri)
	}
}

func diagnosticAllDocs(ctx *Ctx) {
	docDiagnostic.Ctx = ctx

	for uri := range root.Docs {
		docDiagnostic.Set(uri)
	}
}

func scheduleDiagnostic(ctx *Ctx, uri Uri) {
	docDiagnostic.Ctx = ctx
	docDiagnostic.Set(uri)
}

type DiagnosticData struct {
	Type    uint8  `json:"type"`
	Surname string `json:"surname"`
	Name    string `json:"name"`
}
