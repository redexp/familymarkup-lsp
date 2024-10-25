package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func SemanticTokensFull(ctx *glsp.Context, params *proto.SemanticTokensParams) (res *proto.SemanticTokens, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	doc, err := OpenDoc(uri)

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

func GetCaptures(root *Node) ([]*QueryCapture, error) {
	caps, err := familymarkup.GetHighlightCaptures(root)

	if err != nil {
		return nil, err
	}

	list := []*QueryCapture{}

	for _, cap := range caps {
		if cap.Node.IsMissing() {
			continue
		}

		list = append(list, cap)
	}

	return list, nil
}
