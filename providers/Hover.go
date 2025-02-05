package providers

import (
	"fmt"
	"strings"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Hover(ctx *Ctx, params *proto.HoverParams) (h *proto.Hover, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	f, m, target, err := getDefinition(uri, &params.Position)

	if err != nil || (f == nil && m == nil) {
		return
	}

	if f == nil && m != nil {
		f = m.Family
	}

	if m != nil && IsNameDef(target.Parent()) {
		return
	}

	name := f.Name
	aliases := f.Aliases

	if m != nil {
		if m.Node.Parent().Kind() == "sources" {
			return
		}

		name = m.Name
		aliases = m.Aliases
	} else if len(aliases) == 0 {
		return
	}

	message := name

	if len(aliases) > 0 {
		message += " (" + strings.Join(aliases, ", ") + ")"
	}

	if m != nil {
		sources := GetClosestSources(m.Node)
		doc, err := TempDoc(f.Uri)

		if err != nil {
			return nil, err
		}

		message += " - " + L("child_of_source", ToString(sources, doc))
	}

	doc, err := TempDoc(uri)

	if err != nil {
		return
	}

	if IsNameRef(target.Parent()) {
		target = target.Parent()
	}

	r, err := doc.NodeToRange(target)

	Debugf("%s", message)

	if err != nil {
		return
	}

	h = &proto.Hover{
		Range: r,
		Contents: proto.MarkupContent{
			Kind:  proto.MarkupKindMarkdown,
			Value: fmt.Sprintf("```fml\n%s\n```", message),
		},
	}

	Debugf(message)

	return
}
