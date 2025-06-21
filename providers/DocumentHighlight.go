package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocumentHighlight(_ *Ctx, params *proto.DocumentHighlightParams) (res []proto.DocumentHighlight, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	def, err := getDefinition(uri, params.Position)

	if err != nil || def == nil {
		return
	}

	res = make([]proto.DocumentHighlight, 0)
	kind := P(proto.DocumentHighlightKindRead)

	add := func(loc fm.Loc) {
		res = append(res, proto.DocumentHighlight{
			Range: LocToRange(loc),
			Kind:  kind,
		})
	}

	nodes := root.NodeRefs[uri]

	switch def.Type {
	case RefTypeName, RefTypeNameSurname:
		for _, ref := range nodes {
			mem := ref.Member

			if mem == nil {
				continue
			}

			if mem == def.Member || (ref.Type == RefTypeOrigin && mem.Origin == def.Member) {
				add(ref.Token.Loc())
			}
		}

	case RefTypeSurname:
		for _, ref := range nodes {
			if ref.Type == RefTypeSurname && ref.Family == def.Family {
				add(ref.Token.Loc())
			}
		}

	case RefTypeOrigin:
		for _, ref := range nodes {
			mem := ref.Member

			if mem == nil {
				continue
			}

			if mem == def.Member || mem == def.Member.Origin {
				add(ref.Token.Loc())
			}
		}
	}

	return
}
