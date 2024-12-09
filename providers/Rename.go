package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func PrepareRename(ctx *Ctx, params *proto.PrepareRenameParams) (res any, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	doc, err := TempDoc(uri)

	if err != nil {
		return
	}

	node, err := doc.GetClosestNodeByPosition(&params.Position)

	if err != nil || node == nil {
		return
	}

	res = proto.DefaultBehavior{
		DefaultBehavior: true,
	}

	if IsFamilyName(node.Parent()) {
		return
	}

	mem := root.GetMemberByUriNode(uri, node)

	if mem != nil {
		return
	}

	return nil, nil
}

func Rename(ctx *Ctx, params *proto.RenameParams) (res *proto.WorkspaceEdit, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	doc, err := TempDoc(uri)

	if err != nil {
		return
	}

	node, err := doc.GetClosestNodeByPosition(&params.Position)

	if err != nil || node == nil {
		return
	}

	res = &proto.WorkspaceEdit{}

	if IsFamilyName(node.Parent()) {
		r, err := doc.NodeToRange(node)

		if err != nil {
			return nil, err
		}

		res.DocumentChanges = []any{
			proto.TextDocumentEdit{
				TextDocument: proto.OptionalVersionedTextDocumentIdentifier{
					TextDocumentIdentifier: proto.TextDocumentIdentifier{
						URI: uri,
					},
				},
				Edits: []any{
					proto.TextEdit{
						Range:   *r,
						NewText: params.NewName,
					},
				},
			},
		}

		if IsUriName(uri, ToString(node, doc)) {
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

		return res, nil
	}

	_, member, _, err := getDefinition(uri, &params.Position)

	if err != nil || member == nil {
		return
	}

	refs := append(member.Refs, &Ref{
		Uri:  member.Family.Uri,
		Node: member.Node,
	})

	tempDocs := make(Docs)
	changes := make(map[proto.DocumentUri][]proto.TextEdit)

	for _, ref := range refs {
		doc, err := tempDocs.Get(ref.Uri)

		if err != nil {
			return nil, err
		}

		edits, exist := changes[ref.Uri]

		if !exist {
			edits = make([]proto.TextEdit, 0)
		}

		node := ToNameNode(ref.Node)

		r, err := doc.NodeToRange(node)

		if err != nil {
			return nil, err
		}

		changes[ref.Uri] = append(edits, proto.TextEdit{
			Range:   *r,
			NewText: params.NewName,
		})
	}

	res.Changes = changes

	return
}
