package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"

	// "github.com/redexp/textdocument"
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

	// t, nodes, err := GetTypeNode(srcDoc, pos)

	// if err != nil {
	// 	return
	// }

	// if t == "surname" || t == "surname-name" {
	// 	root.UpdateDirty()
	// 	family = root.FindFamily(ToString(nodes[0], srcDoc))

	// 	if family == nil {
	// 		return
	// 	}

	// 	target = family.Node

	// 	if t == "surname-name" {
	// 		member = family.GetMember(ToString(nodes[1], srcDoc))

	// 		if member == nil {
	// 			return
	// 		}

	// 		target = member.Node
	// 	}
	// } else if t == "name" {
	// 	root.UpdateDirty()
	// 	list := root.FindFamiliesByUri(uri)

	// 	if len(list) == 0 {
	// 		return
	// 	}

	// 	familyNode := GetClosestFamilyName(nodes[0])

	// 	if familyNode == nil {
	// 		return
	// 	}

	// 	start := familyNode.StartPoint()
	// 	end := familyNode.EndPoint()

	// 	for _, item := range list {
	// 		if textdocument.CompareNodeWithRange(item.Node, &start, &end) == 0 {
	// 			family = item
	// 			break
	// 		}
	// 	}

	// 	if family == nil {
	// 		return
	// 	}

	// 	member = family.GetMember(ToString(nodes[0], srcDoc))

	// 	if member == nil {
	// 		return
	// 	}

	// 	target = member.Node
	// }

	return
}
