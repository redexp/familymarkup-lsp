package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func FoldingRange(ctx *Ctx, params *proto.FoldingRangeParams) (res []proto.FoldingRange, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	tree := GetTree(uri)

	if tree == nil {
		return
	}

	q, err := CreateQuery(`
		(family) @family

		(relation) @rel
	`)

	if err != nil {
		return
	}

	defer q.Close()

	res = make([]proto.FoldingRange, 0)

	for _, node := range QueryIter(q, tree.RootNode()) {
		start := node.StartPoint().Row
		end := node.EndPoint().Row

		if start == end {
			continue
		}

		res = append(res, proto.FoldingRange{
			StartLine: start,
			EndLine:   end,
			Kind:      P("region"),
		})
	}

	return
}
