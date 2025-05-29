package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func PrepareRename(_ *Ctx, params *proto.PrepareRenameParams) (res any, err error) {
	famMem := root.GetFamMemByPosition(params.TextDocument.URI, params.Position)

	if famMem == nil {
		return
	}

	res = proto.DefaultBehavior{
		DefaultBehavior: true,
	}

	return
}

func Rename(_ *Ctx, params *proto.RenameParams) (res *proto.WorkspaceEdit, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	doc, err := TempDoc(uri)

	if err != nil {
		return
	}

	token := doc.GetTokenByPosition(&params.Position)

	if token == nil {
		return
	}

	res = &proto.WorkspaceEdit{}

	f := doc.FindFamilyByLoc(token.Loc())

	if f != nil && f.Name.IsEqual(token) {
		res.DocumentChanges = []any{
			proto.TextDocumentEdit{
				TextDocument: proto.OptionalVersionedTextDocumentIdentifier{
					TextDocumentIdentifier: proto.TextDocumentIdentifier{
						URI: uri,
					},
				},
				Edits: []any{
					proto.TextEdit{
						Range:   TokenToRange(token),
						NewText: params.NewName,
					},
				},
			},
		}

		if IsUriName(uri, token.Text) {
			newUri, err := RenameUri(uri, params.NewName)

			if err != nil {
				return nil, err
			}

			res.DocumentChanges = append(res.DocumentChanges, proto.RenameFile{
				Kind:   "rename",
				OldURI: uri,
				NewURI: newUri,
			})
		}

		return
	}

	fa, err := getDefinition(uri, params.Position)

	if err != nil || fa == nil || fa.Member == nil {
		return
	}

	member := fa.Member

	refs := append(member.Refs, &Ref{
		Uri:    member.Family.Uri,
		Person: member.Person,
	})

	changes := make(map[proto.DocumentUri][]proto.TextEdit)

	for _, ref := range refs {
		edits, exist := changes[ref.Uri]

		if !exist {
			edits = make([]proto.TextEdit, 0)
		}

		changes[ref.Uri] = append(edits, proto.TextEdit{
			Range:   LocToRange(ref.Person.Loc),
			NewText: params.NewName,
		})
	}

	res.Changes = changes

	return
}
