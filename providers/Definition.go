package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"

	proto "github.com/tliron/glsp/protocol_3_16"
)

func Definition(ctx *Ctx, params *proto.DefinitionParams) (res any, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	family, member, target, err := getDefinition(uri, &params.Position)

	if err != nil {
		return
	}

	doc, err := TempDoc(uri)

	if err != nil {
		return
	}

	var node *Node

	if family != nil {
		node = family.Node
	} else if member != nil {
		node = member.Node
	}

	if node == nil || node == target {
		return
	}

	r, err := doc.NodeToRange(node)

	if err != nil {
		return
	}

	return proto.Location{
		URI:   uri,
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

	target, err = srcDoc.GetClosestNodeByPosition(pos)

	if err != nil || target == nil {
		return
	}

	famMem := root.GetFamMem(uri, target)

	if famMem == nil {
		return
	}

	family = famMem.Family
	member = famMem.Member

	return
}
