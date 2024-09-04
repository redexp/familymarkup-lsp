package src

import (
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func PrepareRename(context *glsp.Context, params *proto.PrepareRenameParams) (res any, err error) {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	doc, err := tempDoc(uri)

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

	if isFamilyName(node.Parent()) {
		return
	}

	mem := root.GetMemberByUriNode(uri, node)

	if mem != nil {
		return
	}

	return nil, nil
}

func Rename(context *glsp.Context, params *proto.RenameParams) (res *proto.WorkspaceEdit, err error) {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	doc, err := tempDoc(uri)

	if err != nil {
		return
	}

	node, err := doc.GetClosestNodeByPosition(&params.Position)

	if err != nil || node == nil {
		return
	}

	res = &proto.WorkspaceEdit{}

	if isFamilyName(node.Parent()) {
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

		if isUriName(uri, toString(node, doc)) {
			newUri, err := renameUri(uri, params.NewName)

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

	family, member, _, err := getDefinition(uri, &params.Position)

	if err != nil || member == nil {
		return
	}

	refs := append(member.Refs, &Ref{
		Uri:  family.Uri,
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

		node := nameRefName(ref.Node)

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
