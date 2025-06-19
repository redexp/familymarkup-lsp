package providers

import (
	. "github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"
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

	res.Changes = make(map[proto.DocumentUri][]proto.TextEdit)

	addMem := func(uri string, person *fm.Person) {
		edits, exist := res.Changes[uri]

		if !exist {
			edits = make([]proto.TextEdit, 0)
		}

		res.Changes[uri] = append(edits, proto.TextEdit{
			Range:   TokenToRange(person.Name),
			NewText: params.NewName,
		})
	}

	addMem(member.Family.Uri, member.Person)

	for uri, person := range member.GetRefsIter() {
		addMem(uri, person)
	}

	return
}
