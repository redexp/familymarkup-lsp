package providers

import (
	"iter"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func References(ctx *Ctx, params *proto.ReferenceParams) (res []proto.Location, err error) {
	family, member, _, err := getDefinition(params.TextDocument.URI, &params.Position)

	if err != nil || (family == nil && member == nil) {
		return
	}

	tempDocs := make(Docs)
	res = make([]proto.Location, 0)

	for uri, node := range GetReferencesIter(family, member) {
		doc, err := tempDocs.Get(uri)

		if err != nil {
			return nil, err
		}

		r, err := doc.NodeToRange(node)

		if err != nil {
			return nil, err
		}

		res = append(res, proto.Location{
			URI:   uri,
			Range: *r,
		})
	}

	if !params.Context.IncludeDeclaration {
		return
	}

	if member != nil && member.InfoUri != "" {
		res = append(res, proto.Location{
			URI: member.InfoUri,
			Range: proto.Range{
				Start: proto.Position{
					Line:      0,
					Character: 0,
				},
				End: proto.Position{
					Line:      0,
					Character: 0,
				},
			},
		})
	}

	return
}

func GetReferencesIter(family *Family, member *Member) iter.Seq2[Uri, *Node] {
	return func(yield func(string, *Node) bool) {
		if family == nil {
			family = &Family{}
		}

		if member == nil {
			member = &Member{}
		}

		for uri, nodes := range root.NodeRefs {
			for node, famMem := range nodes {
				if famMem.Family == family || famMem.Member == member {
					if !yield(uri, node) {
						return
					}
				}
			}
		}
	}
}
