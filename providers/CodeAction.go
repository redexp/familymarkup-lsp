package providers

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

type CodeActionData struct {
	Uri  string `json:"uri"`
	Type uint8  `json:"type"`
	Mod  uint8  `json:"mod"`
	Name string `json:"name"`
}

const (
	CreateFamilyAfterCurrentFamily = iota
	CreateFamilyAtTheEndOfFile
	CreateFamilyOnNewFile
)

func CodeAction(ctx *Ctx, params *proto.CodeActionParams) (any, error) {
	if len(params.Context.Diagnostics) == 0 {
		return nil, nil
	}

	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return nil, err
	}

	list := make([]proto.CodeAction, 0)
	tempDocs := make(Docs)

	for _, d := range params.Context.Diagnostics {
		if d.Data == nil {
			continue
		}

		var data DiagnosticData
		err := mapstructure.Decode(d.Data, &data)

		if err != nil {
			return nil, err
		}

		switch data.Type {
		case UnknownFamilyError:
			doc, err := tempDocs.Get(uri)

			if err != nil {
				return nil, err
			}

			node, err := doc.GetClosestNodeByPosition(&d.Range.Start)

			if err != nil {
				return nil, err
			}

			name := ToString(node, doc)

			family := GetClosestFamilyName(node)

			list = append(
				list,
				proto.CodeAction{
					Title:       L("create_family_after", name, ToString(family, doc)),
					Kind:        P(proto.CodeActionKindQuickFix),
					Diagnostics: []proto.Diagnostic{d},
					Data: CodeActionData{
						Uri:  uri,
						Type: UnknownFamilyError,
						Mod:  CreateFamilyAfterCurrentFamily,
					},
				},
				proto.CodeAction{
					Title:       L("create_family_at_end", name),
					Kind:        P(proto.CodeActionKindQuickFix),
					Diagnostics: []proto.Diagnostic{d},
					Data: CodeActionData{
						Uri:  uri,
						Type: UnknownFamilyError,
						Mod:  CreateFamilyAtTheEndOfFile,
					},
				},
				proto.CodeAction{
					Title:       L("create_family_file", name),
					Kind:        P(proto.CodeActionKindQuickFix),
					Diagnostics: []proto.Diagnostic{d},
					Data: CodeActionData{
						Uri:  uri,
						Type: UnknownFamilyError,
						Mod:  CreateFamilyOnNewFile,
					},
				},
			)

		case NameDuplicateWarning:
			family := root.Families[data.Surname]

			if family == nil {
				continue
			}

			dups := family.Duplicates[data.Name]

			if dups == nil {
				continue
			}

			member := family.Members[data.Name]

			if member == nil {
				continue
			}

			dups = append(dups, &Duplicate{Member: member})

			for _, dup := range dups {
				name := dup.Member.GetUniqName()

				if name == "" {
					continue
				}

				sources := GetClosestSources(dup.Member.Node)

				if sources == nil {
					continue
				}

				doc, err := tempDocs.Get(family.Uri)

				if err != nil {
					return nil, err
				}

				list = append(list, proto.CodeAction{
					Title:       L("change_name_from_source", name, ToString(sources, doc)),
					Kind:        P(proto.CodeActionKindQuickFix),
					Diagnostics: []proto.Diagnostic{d},
					Data: CodeActionData{
						Uri:  uri,
						Type: NameDuplicateWarning,
						Name: name,
					},
				})
			}
		}
	}

	return list, nil
}

func CodeActionResolve(ctx *Ctx, params *proto.CodeAction) (res *proto.CodeAction, err error) {
	if len(params.Diagnostics) == 0 || params.Data == nil {
		return
	}

	var data CodeActionData

	err = mapstructure.Decode(params.Data, &data)

	if err != nil {
		return
	}

	start := params.Diagnostics[0].Range.Start
	end := params.Diagnostics[0].Range.End

	res = &proto.CodeAction{
		Edit: &proto.WorkspaceEdit{},
	}

	switch data.Type {
	case UnknownFamilyError:
		doc, err := TempDoc(data.Uri)

		if err != nil {
			return nil, err
		}

		node, err := doc.GetClosestNodeByPosition(&start)

		if err != nil || node == nil {
			return nil, err
		}

		surname := ToString(node, doc)

		text := fmt.Sprintf("%s\n\n", surname)

		if node.Kind() == "surname" {
			next := node.NextSibling()

			if next != nil && next.Kind() == "name" {
				text = fmt.Sprintf("%s? + ? =\n1. %s", text, ToString(next, doc))
			}
		}

		if data.Mod == CreateFamilyOnNewFile {
			newUri, err := RenameUri(data.Uri, surname)

			if err != nil {
				return nil, err
			}

			createFile := proto.CreateFile{
				Kind: "create",
				URI:  newUri,
			}

			pos := Position{
				Line:      0,
				Character: 0,
			}

			res.Edit.DocumentChanges = []any{
				createFile,
				createInserText(newUri, pos, text),
			}

			scheduleDiagnostic(ctx, data.Uri, nil)

			return res, nil
		}

		var root *Node

		switch data.Mod {
		case CreateFamilyAfterCurrentFamily:
			root = GetClosestNode(node, "family")

		case CreateFamilyAtTheEndOfFile:
			root = doc.Tree.RootNode()
		}

		pos, err := doc.PointToPosition(root.EndPosition())

		if err != nil {
			return nil, err
		}

		res.Edit.DocumentChanges = []any{createInserText(data.Uri, *pos, "\n\n"+text)}

	case NameDuplicateWarning:
		res.Edit.DocumentChanges = []any{createEdit(data.Uri, start, end, data.Name)}
	}

	return res, nil
}

func createEdit(uri Uri, start proto.Position, end proto.Position, text string) proto.TextDocumentEdit {
	return proto.TextDocumentEdit{
		TextDocument: proto.OptionalVersionedTextDocumentIdentifier{
			TextDocumentIdentifier: proto.TextDocumentIdentifier{URI: uri},
		},
		Edits: []any{
			proto.TextEdit{
				Range: proto.Range{
					Start: start,
					End:   end,
				},
				NewText: text,
			},
		},
	}
}

func createInserText(uri Uri, pos proto.Position, text string) proto.TextDocumentEdit {
	return createEdit(uri, pos, pos, text)
}
