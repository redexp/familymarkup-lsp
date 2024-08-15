package src

import (
	"fmt"
	"strings"

	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Hover(context *glsp.Context, params *proto.HoverParams) (h *proto.Hover, err error) {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	f, m, t, doc, err := getDefinition(uri, &params.Position)

	if err != nil || f == nil || t == nil {
		return
	}

	name := f.Name
	aliases := f.Aliases

	if m != nil {
		name = m.Name
		aliases = m.Aliases
	}

	if len(aliases) == 0 {
		return
	}

	r, err := doc.NodeToRange(t)

	if err != nil {
		return
	}

	h = &proto.Hover{
		Range: r,
		Contents: proto.MarkupContent{
			Kind:  proto.MarkupKindMarkdown,
			Value: fmt.Sprintf("**%s** (%s)", name, strings.Join(aliases, ", ")),
		},
	}

	return
}
