package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func References(_ *Ctx, params *proto.ReferenceParams) (res []proto.Location, err error) {
	def, err := getDefinition(params.TextDocument.URI, params.Position)

	if err != nil || def == nil {
		return
	}

	f, mem, _ := def.Spread()

	res = make([]proto.Location, 0)

	switch def.Type {
	case RefTypeName, RefTypeNameSurname, RefTypeOrigin:
		for ref, uri := range mem.GetAllRefsIter() {
			res = append(res, proto.Location{
				URI:   uri,
				Range: TokenToRange(ref.Token),
			})
		}

	case RefTypeSurname:
		for ref, uri := range f.GetRefsIter() {
			res = append(res, proto.Location{
				URI:   uri,
				Range: TokenToRange(ref.Token),
			})
		}
	}

	if !params.Context.IncludeDeclaration {
		return
	}

	if mem != nil && mem.InfoUri != "" {
		res = append(res, proto.Location{
			URI: mem.InfoUri,
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
