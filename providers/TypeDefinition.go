package providers

import (
	proto "github.com/tliron/glsp/protocol_3_16"
)

func TypeDefinition(_ *Ctx, params *proto.TypeDefinitionParams) (res any, err error) {
	_, mem, _, err := getDefinition(params.TextDocument.URI, &params.Position)

	if err != nil || mem == nil || mem.InfoUri == "" {
		return
	}

	return proto.Location{
		URI: mem.InfoUri,
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
