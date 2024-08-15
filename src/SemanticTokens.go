package src

import (
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func SemanticTokensFull(ctx *glsp.Context, params *proto.SemanticTokensParams) (res *proto.SemanticTokens, err error) {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	doc, err := openDoc(uri)

	if err != nil {
		return
	}

	tokens, err := doc.ConvertHighlightCaptures(typesMap)

	if err != nil {
		return
	}

	res = &proto.SemanticTokens{
		Data: tokens,
	}

	return res, nil
}

func GetCaptures(root *sitter.Node) ([]*sitter.QueryCapture, error) {
	caps, err := familymarkup.GetHighlightCaptures(root)

	if err != nil {
		return nil, err
	}

	list := []*sitter.QueryCapture{}

	for _, cap := range caps {
		if cap.Node.IsMissing() {
			continue
		}

		list = append(list, cap)
	}

	return list, nil
}
