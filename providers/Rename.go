package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func PrepareRename(_ *Ctx, params *proto.PrepareRenameParams) (res any, err error) {
	def, err := getDefinition(params.TextDocument.URI, params.Position)

	if err != nil || def == nil {
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

	def, err := getDefinition(uri, params.Position)

	if err != nil || def == nil {
		return
	}

	res = &proto.WorkspaceEdit{}

	if def.Type == RefTypeSurname {
		f := def.Family

		edits := make([]any, 0)

		for ref, uri := range f.GetRefsIter() {
			edits = append(edits, proto.TextDocumentEdit{
				TextDocument: proto.OptionalVersionedTextDocumentIdentifier{
					TextDocumentIdentifier: proto.TextDocumentIdentifier{
						URI: uri,
					},
				},
				Edits: []any{
					proto.TextEdit{
						Range:   TokenToRange(ref.Token),
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

	member := def.Member

	if member == nil {
		return
	}

	res.Changes = make(map[proto.DocumentUri][]proto.TextEdit)

	for ref, refUri := range member.GetAllRefsIter() {
		edits, ok := res.Changes[refUri]

		if !ok {
			edits = make([]proto.TextEdit, 0)
		}

		res.Changes[refUri] = append(edits, proto.TextEdit{
			Range:   TokenToRange(ref.Person.Name),
			NewText: params.NewName,
		})
	}

	return
}
