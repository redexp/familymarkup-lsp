package src

import (
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func FoldingRange(context *glsp.Context, params *proto.FoldingRangeParams) (res []proto.FoldingRange, err error) {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	tree := getTree(uri)

	if tree == nil {
		return
	}

	q, err := createQuery(`
		(family) @family

		(relation) @rel
	`)

	if err != nil {
		return
	}

	defer q.Close()

	res = make([]proto.FoldingRange, 0)

	for _, node := range queryIter(q, tree.RootNode()) {
		start := node.StartPoint().Row
		end := node.EndPoint().Row

		if start == end {
			continue
		}

		res = append(res, proto.FoldingRange{
			StartLine: start,
			EndLine:   end,
			Kind:      pt("region"),
		})
	}

	return
}
