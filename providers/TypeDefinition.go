package providers

import (
	proto "github.com/tliron/glsp/protocol_3_16"
)

func TypeDefinition(_ *Ctx, params *proto.TypeDefinitionParams) (res any, err error) {
	fa, err := getDefinition(params.TextDocument.URI, params.Position)

	if err != nil || fa == nil || fa.Member == nil || fa.Member.InfoUri == "" {
		return
	}

	return proto.Location{
		URI: fa.Member.InfoUri,
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
	}, nil
}
