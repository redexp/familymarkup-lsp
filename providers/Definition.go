package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"

	proto "github.com/tliron/glsp/protocol_3_16"
)

func Definition(ctx *Ctx, params *proto.DefinitionParams) (res any, err error) {
	family, _, target, err := getDefinition(params.TextDocument.URI, &params.Position)

	if err != nil || target == nil {
		return
	}

	doc, err := TempDoc(family.Uri)

	if err != nil {
		return
	}

	r, err := doc.NodeToRange(target)

	if err != nil {
		return
	}

	return proto.Location{
		URI:   family.Uri,
		Range: *r,
	}, nil
}

func getDefinition(uri Uri, pos *Position) (family *Family, member *Member, target *Node, err error) {
	uri, err = NormalizeUri(uri)

	if err != nil {
		return
	}

	err = root.UpdateDirty()

	if err != nil {
		LogDebug("getDefinition UpdateDirty %s", err)
	}

	srcDoc, err := TempDoc(uri)

	if err != nil {
		return
	}

	node, err := srcDoc.GetClosestNodeByPosition(pos)

	if err != nil || node == nil {
		return
	}

	member = root.GetMemberByUriNode(uri, node)

	if member != nil {
		return member.Family, member, member.Node, nil
	}

	return
}
