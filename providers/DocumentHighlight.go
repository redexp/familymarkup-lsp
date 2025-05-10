package providers

import (
	. "github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocumentHighlight(_ *Ctx, params *proto.DocumentHighlightParams) (res []proto.DocumentHighlight, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	nodes, exist := root.NodeRefs[uri]

	if !exist {
		return
	}

	family, member, _, err := getDefinition(uri, params.Position)

	if err != nil {
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

	if family != nil && family.Uri == uri {
		add(family.Node.Loc)
	}

	if member != nil && member.Family.Uri == uri {
		add(member.Person.Loc)
	}

	for _, famMem := range nodes {
		if (family == nil || famMem.Family != family) && (member == nil || famMem.Member != member) {
			continue
		}

		add(famMem.Token.Loc())
	}

	return
}
