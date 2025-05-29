package providers

import (
	"fmt"
	fm "github.com/redexp/familymarkup-parser"
	"strings"

	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Hover(_ *Ctx, params *proto.HoverParams) (h *proto.Hover, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	fa, err := getDefinition(uri, params.Position)

	if err != nil || fa == nil {
		return
	}

	f, m, target := fa.Spread()

	if f == nil {
		f = m.Family
	}

	if m != nil && fa.Person != nil && fa.Person.IsChild {
		return
	}

	name := f.Name
	aliases := f.Aliases

	if m != nil {
		if m.Person.Side == fm.SideSources {
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
		sources := m.Person.Relation.Sources

		message += " - " + L("child_of_source", sources.Format())
	}

	r := TokenToRange(target)

	if fa.Person != nil && !fa.Person.IsChild {
		r = LocToRange(fa.Person.Loc)
	}

	h = &proto.Hover{
		Range: &r,
		Contents: proto.MarkupContent{
			Kind:  proto.MarkupKindMarkdown,
			Value: fmt.Sprintf("```fml\n%s\n```", message),
		},
	}

	return
}
