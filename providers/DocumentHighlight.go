package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocumentHighlight(ctx *Ctx, params *proto.DocumentHighlightParams) (res []proto.DocumentHighlight, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	nodes, exist := root.NodeRefs[uri]

	if !exist {
		return
	}

	family, member, _, err := getDefinition(uri, &params.Position)

	if err != nil {
		return
	}

	doc, err := TempDoc(uri)

	if err != nil {
		return
	}

	res = make([]proto.DocumentHighlight, 0)
	kind := P(proto.DocumentHighlightKindRead)

	add := func(node *Node) error {
		r, err := doc.NodeToRange(node)

		if err != nil {
			return err
		}

		res = append(res, proto.DocumentHighlight{
			Range: *r,
			Kind:  kind,
		})

		return nil
	}

	if family != nil && family.Uri == uri {
		add(family.Node)
	}

	if member != nil && member.Family.Uri == uri {
		add(member.Node)
	}

	for _, famMem := range nodes {
		if (family == nil || famMem.Family != family) && (member == nil || famMem.Member != member) {
			continue
		}

		err = add(famMem.Node)

		if err != nil {
			return
		}
	}

	return
}
