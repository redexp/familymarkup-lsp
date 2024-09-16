package providers

import (
	"fmt"
	"strings"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Hover(context *glsp.Context, params *proto.HoverParams) (h *proto.Hover, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	f, m, _, err := getDefinition(uri, &params.Position)

	if err != nil || f == nil {
		return
	}

	name := f.Name
	aliases := f.Aliases

	if m != nil {
		if m.Node.Parent().Type() == "sources" {
			return
		}

		name = m.Name
		aliases = m.Aliases
	} else if len(aliases) == 0 {
		return
	}

	message := fmt.Sprintf("**%s**", name)

	if len(aliases) > 0 {
		message += " (" + strings.Join(aliases, ", ") + ")"
	}

	if m != nil {
		sources := GetClosestSources(m.Node)
		doc, err := TempDoc(f.Uri)

		if err != nil {
			return nil, err
		}

		message += " child of " + ToString(sources, doc)
	}

	doc, err := TempDoc(uri)

	if err != nil {
		return
	}

	target, err := doc.GetClosestNodeByPosition(&params.Position)

	if err != nil || target == nil {
		return
	}

	if IsNameRef(target.Parent()) {
		target = target.Parent()
	}

	r, err := doc.NodeToRange(target)

	if err != nil {
		return
	}

	h = &proto.Hover{
		Range: r,
		Contents: proto.MarkupContent{
			Kind:  proto.MarkupKindMarkdown,
			Value: message,
		},
	}

	return
}
