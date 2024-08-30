package src

import (
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocumentHighlight(context *glsp.Context, params *proto.DocumentHighlightParams) (res []proto.DocumentHighlight, err error) {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	family, member, _, err := getDefinition(uri, &params.Position)

	if err != nil || member == nil || len(member.Refs) == 0 {
		return
	}

	doc, err := openDoc(uri)

	if err != nil {
		return
	}

	res = make([]proto.DocumentHighlight, 0)

	if family.Uri == uri {
		r, err := doc.NodeToRange(member.Node)

		if err != nil {
			return nil, err
		}

		res = append(res, proto.DocumentHighlight{
			Range: *r,
			Kind:  pt(proto.DocumentHighlightKindRead),
		})
	}

	for _, ref := range member.Refs {
		if ref.Uri != uri {
			continue
		}

		r, err := doc.NodeToRange(ref.Node)

		if err != nil {
			return nil, err
		}

		res = append(res, proto.DocumentHighlight{
			Range: *r,
			Kind:  pt(proto.DocumentHighlightKindRead),
		})
	}

	return res, nil
}
