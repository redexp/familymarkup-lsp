package providers

import (
	"encoding/json"
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"
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

func TextDocumentDiagnostic(_ *Ctx, params *DocumentDiagnosticParams) (res *DocumentDiagnosticReport, err error) {
	err = root.UpdateDirty()

	if err != nil {
		return
	}

	uri := NormalizeUri(params.TextDocument.URI)
	doc := GetDoc(uri)

	if !doc.NeedDiagnostic {
		res = &DocumentDiagnosticReport{
			Kind:     "unchanged",
			ResultId: doc.V(),
		}
		return
	}

	doc.NeedDiagnostic = false

	res = &DocumentDiagnosticReport{
		Kind:     "full",
		ResultId: doc.V(),
		Items:    GetDiagnostics(uri),
	}

	return
}

func WorkspaceDiagnostic(_ *Ctx, _ *WorkspaceDiagnosticParams) (res *WorkspaceDiagnosticReport, err error) {
	err = root.UpdateDirty()

	if err != nil {
		return
	}

	count := len(root.Docs)

	res = &WorkspaceDiagnosticReport{
		Items: make([]WorkspaceDocumentDiagnosticReport, count),
	}

	i := 0

	for uri, doc := range root.Docs {
		if doc.NeedDiagnostic {
			doc.NeedDiagnostic = false

			res.Items[i] = WorkspaceDocumentDiagnosticReport{
				Kind:     "full",
				Uri:      uri,
				ResultId: doc.V(),
				Items:    GetDiagnostics(uri),
			}
		} else {
			res.Items[i] = WorkspaceDocumentDiagnosticReport{
				Kind:     "unchanged",
				Uri:      uri,
				ResultId: doc.V(),
			}
		}

		i++
	}

	return
}

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

	// syntax errors
	for _, token := range doc.Tokens {
		if token.Type == fm.TokenInvalid || token.ErrType == fm.ErrUnexpected {
			add(proto.Diagnostic{
				Severity: Error,
				Range:    TokenToRange(token),
				Message:  L("syntax_error"),
			})
		}
	}

	// unknown refs
	for _, ref := range root.UnknownRefs {
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

type DiagnosticData struct {
	Type    uint8  `json:"type"`
	Surname string `json:"surname"`
	Name    string `json:"name"`
}

type DiagnosticHandler struct {
	TextDocumentDiagnostic TextDocumentDiagnosticFunc
	WorkspaceDiagnostic    WorkspaceDiagnosticFunc
}

func (req *DiagnosticHandler) Handle(ctx *Ctx) (res any, validMethod bool, validParams bool, err error) {
	switch ctx.Method {
	case TextDocumentDiagnosticMethod:
		validMethod = true

		var params DocumentDiagnosticParams
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.TextDocumentDiagnostic(ctx, &params)
		}

	case WorkspaceDiagnosticMethod:
		validMethod = true

		var params WorkspaceDiagnosticParams
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.WorkspaceDiagnostic(ctx, &params)
		}
	}

	return
}

const TextDocumentDiagnosticMethod = "textDocument/diagnostic"

type TextDocumentDiagnosticFunc func(ctx *Ctx, params *DocumentDiagnosticParams) (*DocumentDiagnosticReport, error)

type DocumentDiagnosticParams struct {
	TextDocument     proto.TextDocumentIdentifier `json:"textDocument"`
	PreviousResultId string                       `json:"previousResultId,omitempty"`
}

type DocumentDiagnosticReport struct {
	Kind             string                           `json:"kind,omitempty"`
	Items            []proto.Diagnostic               `json:"items"`
	ResultId         string                           `json:"resultId,omitempty"`
	RelatedDocuments map[Uri]DocumentDiagnosticReport `json:"relatedDocuments,omitempty"`
}

type DiagnosticOptions struct {
	InterFileDependencies bool `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool `json:"workspaceDiagnostics"`
}

const WorkspaceDiagnosticMethod = "workspace/diagnostic"

type WorkspaceDiagnosticFunc func(ctx *Ctx, params *WorkspaceDiagnosticParams) (*WorkspaceDiagnosticReport, error)

type WorkspaceDiagnosticParams struct {
	PreviousResultIds []PreviousResultId `json:"previousResultIds"`
}

type PreviousResultId struct {
	Uri   string `json:"uri"`
	Value string `json:"value"`
}

type WorkspaceDiagnosticReport struct {
	Items []WorkspaceDocumentDiagnosticReport `json:"items"`
}

type WorkspaceDocumentDiagnosticReport struct {
	Kind     string             `json:"kind,omitempty"`
	Uri      string             `json:"uri"`
	ResultId string             `json:"resultId,omitempty"`
	Version  int                `json:"version"`
	Items    []proto.Diagnostic `json:"items"`
}
