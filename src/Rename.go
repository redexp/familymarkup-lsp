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

	_, member, _, _, err := getDefinition(uri, &params.Position)

	if err != nil || member == nil {
		return
	}

	res = proto.DefaultBehavior{
		DefaultBehavior: true,
	}

	return
}

func Rename(context *glsp.Context, params *proto.RenameParams) (res *proto.WorkspaceEdit, err error) {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	family, member, _, _, err := getDefinition(uri, &params.Position)

	if err != nil || member == nil {
		return
	}

	refs := append(member.Refs, &Ref{
		Uri:  family.Uri,
		Node: member.Node,
	})

	changes := make(map[proto.DocumentUri][]proto.TextEdit)

	for _, ref := range refs {
		doc, err := openDoc(ref.Uri)

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

	res = &proto.WorkspaceEdit{
		Changes: changes,
	}

	return
}
