package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func References(ctx *Ctx, params *proto.ReferenceParams) (res []proto.Location, err error) {
	family, member, target, err := getDefinition(params.TextDocument.URI, &params.Position)

	if err != nil || member == nil {
		return
	}

	tempDocs := make(Docs)
	res = make([]proto.Location, len(member.Refs))

	for i, ref := range member.Refs {
		doc, err := tempDocs.Get(ref.Uri)

		if err != nil {
			return nil, err
		}

		r, err := doc.NodeToRange(ref.Node)

		if err != nil {
			return nil, err
		}

		res[i] = proto.Location{
			URI:   ref.Uri,
			Range: *r,
		}
	}

	if !params.Context.IncludeDeclaration {
		return
	}

	if member.InfoUri != "" {
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

	doc, err := tempDocs.Get(family.Uri)

	if err != nil {
		return
	}

	r, err := doc.NodeToRange(target)

	if err != nil {
		return
	}

	res = append(res, proto.Location{
		URI:   family.Uri,
		Range: *r,
	})

	return
}
