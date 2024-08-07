package src

import (
	"github.com/redexp/textdocument"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Definition(context *glsp.Context, params *proto.DefinitionParams) (res any, err error) {
	family, _, target, doc, err := getDefinition(params.TextDocument.URI, &params.Position)

	if err != nil || target == nil {
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

func getDefinition(uri Uri, pos *Position) (family *Family, member *Member, target *Node, doc *TextDocument, err error) {
	srcDoc, err := openDoc(uri)

	if err != nil {
		return
	}

	t, nodes, err := getTypeNode(srcDoc, pos)

	if err != nil {
		return
	}

	if t == "surname" || t == "surname-name" {
		family = root.FindFamily(toString(nodes[0], srcDoc))

		if family == nil {
			return
		}

		target = family.Node

		if t == "surname-name" {
			member = family.FindMember(toString(nodes[1], srcDoc))

			if member == nil {
				return
			}

			target = member.Node
		}
	} else if t == "name" {
		list := root.FindFamiliesByUri(uri)

		if len(list) == 0 {
			return
		}

		familyNode := getClosestFamilyName(nodes[0])

		if familyNode == nil {
			return
		}

		start := familyNode.StartPoint()
		end := familyNode.EndPoint()

		for _, item := range list {
			if textdocument.CompareNodeWithRange(item.Node, &start, &end) == 0 {
				family = item
				break
			}
		}

		if family == nil {
			return
		}

		member = family.FindMember(toString(nodes[0], srcDoc))

		if member == nil {
			return
		}

		target = member.Node
	}

	if family != nil {
		doc, err = openDoc(family.Uri)
	}

	return
}
