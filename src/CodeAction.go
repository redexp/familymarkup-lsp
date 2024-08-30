package src

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

type CodeActionData struct {
	Uri  string `json:"uri"`
	Type uint8  `json:"type"`
	Mod  uint8  `json:"mod"`
}

const (
	CreateFamilyAfterCurrentFamily = iota
	CreateFamilyAtTheEndOfFile
	CreateFamilyOnNewFile
)

func CodeAction(context *glsp.Context, params *proto.CodeActionParams) (any, error) {
	if len(params.Context.Diagnostics) == 0 {
		return nil, nil
	}

	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return nil, err
	}

	list := make([]proto.CodeAction, 0)
	kindQuickFix := proto.CodeActionKindQuickFix

	for _, d := range params.Context.Diagnostics {
		data, ok := d.Data.(float64)

		if !ok {
			continue
		}

		switch data {
		case UnknownFamilyError:
			doc, err := tempDoc(uri)

			if err != nil {
				return nil, err
			}

			node, err := doc.GetClosestNodeByPosition(&d.Range.Start)

			if err != nil {
				return nil, err
			}

			name := toString(node, doc)

			family := getClosestFamilyName(node)

			list = append(
				list,
				proto.CodeAction{
					Title:       fmt.Sprintf("Create %s family after %s", name, toString(family, doc)),
					Kind:        &kindQuickFix,
					Diagnostics: []proto.Diagnostic{d},
					Data: CodeActionData{
						Uri:  uri,
						Type: UnknownFamilyError,
						Mod:  CreateFamilyAfterCurrentFamily,
					},
				},
				proto.CodeAction{
					Title:       fmt.Sprintf("Create %s family at the end of file", name),
					Kind:        &kindQuickFix,
					Diagnostics: []proto.Diagnostic{d},
					Data: CodeActionData{
						Uri:  uri,
						Type: UnknownFamilyError,
						Mod:  CreateFamilyAtTheEndOfFile,
					},
				},
				proto.CodeAction{
					Title:       fmt.Sprintf("Create new file with %s family", name),
					Kind:        &kindQuickFix,
					Diagnostics: []proto.Diagnostic{d},
					Data: CodeActionData{
						Uri:  uri,
						Type: UnknownFamilyError,
						Mod:  CreateFamilyOnNewFile,
					},
				},
			)
		}
	}

	return list, nil
}

func CodeActionResolve(ctx *glsp.Context, params *proto.CodeAction) (res *proto.CodeAction, err error) {
	if len(params.Diagnostics) == 0 || params.Data == nil {
		return
	}

	var data CodeActionData

	err = mapstructure.Decode(params.Data, &data)

	if err != nil {
		return
	}

	start := params.Diagnostics[0].Range.Start

	switch data.Type {
	case UnknownFamilyError:
		doc, err := tempDoc(data.Uri)

		if err != nil {
			return nil, err
		}

		node, err := doc.GetClosestNodeByPosition(&start)

		if err != nil || node == nil {
			return nil, err
		}

		surname := toString(node, doc)

		text := fmt.Sprintf("%s\n\n", surname)

		if node.Type() == "surname" {
			next := node.NextSibling()

			if next != nil && next.Type() == "name" {
				text = fmt.Sprintf("%s? + ? =\n1. %s", text, toString(next, doc))
			}
		}

		res = &proto.CodeAction{
			Edit: &proto.WorkspaceEdit{},
		}

		if data.Mod == CreateFamilyOnNewFile {
			newUri, err := renameUri(data.Uri, surname)

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
				createEdit(newUri, pos, text),
			}

			docDiagnostic.Set(data.Uri, ctx)

			return res, nil
		}

		var root *Node

		switch data.Mod {
		case CreateFamilyAfterCurrentFamily:
			root = getClosestNode(node, "family")

		case CreateFamilyAtTheEndOfFile:
			root = doc.Tree.RootNode()
		}

		pos, err := doc.PointToPosition(root.EndPoint())

		if err != nil {
			return nil, err
		}

		res.Edit.DocumentChanges = []any{createEdit(data.Uri, *pos, "\n\n"+text)}
	}

	return res, nil
}

func createEdit(uri Uri, pos proto.Position, text string) proto.TextDocumentEdit {
	return proto.TextDocumentEdit{
		TextDocument: proto.OptionalVersionedTextDocumentIdentifier{
			TextDocumentIdentifier: proto.TextDocumentIdentifier{URI: uri},
		},
		Edits: []any{
			proto.TextEdit{
				Range: proto.Range{
					Start: pos,
					End:   pos,
				},
				NewText: text,
			},
		},
	}
}
