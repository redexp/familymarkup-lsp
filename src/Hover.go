package src

import (
	"fmt"
	"strings"

	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Hover(context *glsp.Context, params *proto.HoverParams) (h *proto.Hover, err error) {
	f, m, t, doc, err := getDefinition(params.TextDocument.URI, &params.Position)

	if err != nil || f == nil || t == nil {
		return
	}

	r, err := doc.NodeToRange(t)

	if err != nil {
		return
	}

	toAliases := func(aliases []string) string {
		if len(aliases) == 0 {
			return ""
		}

		return " (" + strings.Join(aliases, ", ") + ")"
	}

	h = &proto.Hover{
		Range: r,
	}

	if m != nil {
		h.Contents = proto.MarkupContent{
			Kind:  proto.MarkupKindMarkdown,
			Value: fmt.Sprintf("**%s** %s", m.Name, toAliases(m.Aliases)),
		}
	} else {
		h.Contents = proto.MarkupContent{
			Kind:  proto.MarkupKindMarkdown,
			Value: fmt.Sprintf("**%s** %s", f.Name, toAliases(f.Aliases)),
		}
	}

	return
}
