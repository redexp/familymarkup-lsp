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

	doc, err := TempDoc(uri)

	if err != nil {
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

	for _, node := range QueryIter(q, doc.Tree.RootNode(), []byte(doc.Text)) {
		start := node.StartPosition().Row
		end := node.EndPosition().Row

		if start == end {
			continue
		}

		res = append(res, proto.FoldingRange{
			StartLine: uint32(start),
			EndLine:   uint32(end),
			Kind:      P("region"),
		})
	}

	return
}
