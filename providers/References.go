package providers

import (
	fm "github.com/redexp/familymarkup-parser"
	"iter"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func References(ctx *Ctx, params *proto.ReferenceParams) (res []proto.Location, err error) {
	family, member, _, err := getDefinition(params.TextDocument.URI, params.Position)

	if err != nil || (family == nil && member == nil) {
		return
	}

	res = make([]proto.Location, 0)

	for uri, loc := range GetReferencesIter(family, member) {
		res = append(res, proto.Location{
			URI:   uri,
			Range: *LocToRange(loc),
		})
	}

	if !params.Context.IncludeDeclaration {
		return
	}

	if member != nil && member.InfoUri != "" {
		res = append(res, proto.Location{
			URI: member.InfoUri,
			Range: Range{
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

func GetReferencesIter(family *Family, member *Member) iter.Seq2[Uri, fm.Loc] {
	return func(yield func(string, fm.Loc) bool) {
		if family == nil {
			family = &Family{}
		}

		if member == nil {
			member = &Member{}
		}

		for uri, nodes := range root.NodeRefs {
			for _, item := range nodes {
				if item.Family == family || item.Member == member {
					if !yield(uri, item.Loc) {
						return
					}
				}
			}
		}
	}
}
