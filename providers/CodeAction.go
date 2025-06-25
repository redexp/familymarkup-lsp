package providers

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"
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

func CodeAction(_ *Ctx, params *proto.CodeActionParams) (res any, err error) {
	if len(params.Context.Diagnostics) == 0 {
		return
	}

	uri := NormalizeUri(params.TextDocument.URI)

	list := make([]proto.CodeAction, 0)
	QuickFix := P(proto.CodeActionKindQuickFix)

	add := func(items ...proto.CodeAction) {
		list = append(list, items...)
	}

	for _, d := range params.Context.Diagnostics {
		if d.Data == nil {
			continue
		}

		var data DiagnosticData
		err = mapstructure.Decode(d.Data, &data)

		if err != nil {
			return
		}

		switch data.Type {
		case UnknownFamilyError:
			doc := GetDoc(uri)
			family := doc.FindFamilyByRange(d.Range)

			if family == nil {
				continue
			}

			token := doc.GetTokenByPosition(d.Range.Start)

			add(
				proto.CodeAction{
					Title:       L("create_family_after", token.Text, family.Name.Text),
					Kind:        QuickFix,
					Diagnostics: []proto.Diagnostic{d},
					Data: CodeActionData{
						Uri:  uri,
						Type: UnknownFamilyError,
						Mod:  CreateFamilyAfterCurrentFamily,
					},
				},
				proto.CodeAction{
					Title:       L("create_family_at_end", token.Text),
					Kind:        QuickFix,
					Diagnostics: []proto.Diagnostic{d},
					Data: CodeActionData{
						Uri:  uri,
						Type: UnknownFamilyError,
						Mod:  CreateFamilyAtTheEndOfFile,
					},
				},
				proto.CodeAction{
					Title:       L("create_family_file", token.Text),
					Kind:        QuickFix,
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

			doc := GetDoc(family.Uri)

			for _, dup := range dups {
				mem := dup.Member
				name := mem.GetUniqName()

				if name == "" {
					continue
				}

				rel := doc.FindRelationByRange(LocToRange(mem.Person.Loc))

				if rel == nil {
					continue
				}

				source := doc.GetTextByLoc(rel.Sources.Loc)

				add(proto.CodeAction{
					Title:       L("change_name_from_source", name, source),
					Kind:        QuickFix,
					Diagnostics: []proto.Diagnostic{d},
					Data: CodeActionData{
						Uri:  uri,
						Type: NameDuplicateWarning,
						Name: name,
					},
				})
			}

		case ChildWithoutRelationsInfo:
			add(proto.CodeAction{
				Title:       L("create_child_relation"),
				Kind:        QuickFix,
				Diagnostics: []proto.Diagnostic{d},
				Data: CodeActionData{
					Uri:  uri,
					Type: ChildWithoutRelationsInfo,
				},
			})
		}
	}

	return list, nil
}

func CodeActionResolve(_ *Ctx, params *proto.CodeAction) (res *proto.CodeAction, err error) {
	if len(params.Diagnostics) == 0 || params.Data == nil {
		return
	}

	var data CodeActionData

	err = mapstructure.Decode(params.Data, &data)

	if err != nil {
		return
	}

	r := params.Diagnostics[0].Range

	var token *fm.Token

	res = &proto.CodeAction{
		Edit: &proto.WorkspaceEdit{},
	}

	switch data.Type {
	case UnknownFamilyError:
		doc := GetDoc(data.Uri)

		token = doc.GetTokenByPosition(r.Start)

		if token == nil {
			return
		}

		surname := token.Text

		text := fmt.Sprintf("%s\n\n", surname)

		person := doc.FindPersonByRange(r)

		if person != nil && !person.IsChild {
			text = fmt.Sprintf("%s? + ? =\n1. %s", text, person.Name.Text)
		} else if person != nil && person.IsChild && person.Surname == token {
			f := doc.FindFamilyByRange(r)
			text = fmt.Sprintf("%s? + %s %s = ", text, person.Name.Text, f.Name.Text)
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
				createInsertText(newUri, pos, text),
			}

			return res, nil
		}

		var pos Position

		switch data.Mod {
		case CreateFamilyAtTheEndOfFile:
			pos = LocPosToPosition(doc.Root.End)

		default:
			f := doc.FindFamilyByRange(r)

			if f != nil {
				pos = LocPosToPosition(f.End)
			}
		}

		res.Edit.DocumentChanges = []any{createInsertText(data.Uri, pos, "\n\n"+text)}

	case NameDuplicateWarning:
		res.Edit.DocumentChanges = []any{createEdit(data.Uri, r.Start, r.End, data.Name)}

	case ChildWithoutRelationsInfo:
		doc := GetDoc(data.Uri)

		token = doc.GetTokenByPosition(r.Start)

		if token == nil {
			return
		}

		f := doc.FindFamilyByRange(r)

		if f == nil {
			return
		}

		pos := LocPosToPosition(f.End)

		res.Edit.DocumentChanges = []any{createInsertText(data.Uri, pos, fmt.Sprintf("\n\n%s + ? =\n", token.Text))}
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

func createInsertText(uri Uri, pos proto.Position, text string) proto.TextDocumentEdit {
	return createEdit(uri, pos, pos, text)
}
