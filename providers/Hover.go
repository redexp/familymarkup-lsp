package providers

import (
	"fmt"
	fm "github.com/redexp/familymarkup-parser"
	"strings"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Hover(_ *Ctx, params *proto.HoverParams) (h *proto.Hover, err error) {
	ref, err := getDefinition(params.TextDocument.URI, params.Position)

	if err != nil || ref == nil {
		return
	}

	f, mem, target := ref.Spread()

	var name string
	var aliases []string
	var surname string
	var sources *fm.RelList

	switch ref.Type {
	case RefTypeName, RefTypeNameSurname:
		if ref.Person.IsChild || mem.Person == ref.Person {
			return
		}

		if ref.Type == RefTypeName && mem.Origin != nil {
			mem = mem.Origin
			surname = mem.Family.Name
		}

		name = mem.Name
		aliases = mem.Aliases
		sources = mem.Person.Relation.Sources

	case RefTypeSurname:
		if ref.Token == f.Node.Name {
			return
		}

		name = f.Name
		aliases = f.Aliases

	case RefTypeOrigin:
		origin := mem.Origin

		name = origin.Name
		aliases = origin.Aliases
		sources = origin.Person.Relation.Sources
	}

	message := name

	if len(aliases) > 0 {
		message += " (" + strings.Join(aliases, ", ") + ")"
	}

	if surname != "" {
		message += " " + surname
	}

	if sources != nil {
		message += " - " + L("child_of_source", sources.Format())
	}

	if message == target.Text {
		return
	}

	r := TokenToRange(target)

	h = &proto.Hover{
		Range: &r,
		Contents: proto.MarkupContent{
			Kind:  proto.MarkupKindMarkdown,
			Value: fmt.Sprintf("```fml\n%s\n```", message),
		},
	}

	return
}
