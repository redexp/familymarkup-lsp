package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocumentHighlight(_ *Ctx, params *proto.DocumentHighlightParams) (res []proto.DocumentHighlight, err error) {
	uri := NormalizeUri(params.TextDocument.URI)

	def, err := getDefinition(uri, params.Position)

	if err != nil {
		return
	}

	kind := P(proto.DocumentHighlightKindRead)

	add := func(loc fm.Loc) {
		res = append(res, proto.DocumentHighlight{
			Range: LocToRange(loc),
			Kind:  kind,
		})
	}

	if def == nil {
		doc := GetDoc(uri)
		token := doc.GetTokenByPosition(params.Position)

		if token != nil && token.Type == fm.TokenUnknown {
			add(token.Loc())
		}

		return
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
