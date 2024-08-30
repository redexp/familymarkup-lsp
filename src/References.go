package src

import (
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func References(context *glsp.Context, params *proto.ReferenceParams) (res []proto.Location, err error) {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	family, member, target, err := getDefinition(uri, &params.Position)

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

	if params.Context.IncludeDeclaration {
		doc, err := tempDocs.Get(uri)

		if err != nil {
			return nil, err
		}

		r, err := doc.NodeToRange(target)

		if err != nil {
			return nil, err
		}

		res = append(res, proto.Location{
			URI:   family.Uri,
			Range: *r,
		})
	}

	return
}