package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func PrepareRename(_ *Ctx, params *proto.PrepareRenameParams) (res any, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	famMem := root.GetFamMemByPosition(uri, params.Position)

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

	famMem := root.GetFamMemByPosition(uri, params.Position)

	if famMem == nil {
		return
	}

	res = &proto.WorkspaceEdit{}

	f := famMem.Family

	if f != nil {
		edits := make([]any, 0)

		for uri, token := range f.GetRefsIter() {
			edits = append(edits, proto.TextDocumentEdit{
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
			})
		}

		if IsUriName(uri, f.Name) {
			newUri, err := RenameUri(uri, params.NewName)

			if err != nil {
				return nil, err
			}

			edits = append(edits, proto.RenameFile{
				Kind:   "rename",
				OldURI: uri,
				NewURI: newUri,
			})
		}

		res.DocumentChanges = edits

		return
	}

	member := famMem.Member

	if member == nil {
		return
	}

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
			Range:   TokenToRange(ref.Person.Name),
			NewText: params.NewName,
		})
	}

	res.Changes = changes

	return
}
