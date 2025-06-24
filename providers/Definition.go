package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"

	proto "github.com/tliron/glsp/protocol_3_16"
)

func Definition(_ *Ctx, params *proto.DefinitionParams) (res any, err error) {
	uri := NormalizeUri(params.TextDocument.URI)

	ref, err := getDefinition(uri, params.Position)

	if err != nil || ref == nil {
		return
	}

	f, mem, source := ref.Spread()

	var target *fm.Token

	switch ref.Type {
	case RefTypeName, RefTypeNameSurname:
		uri = mem.Family.Uri
		target = mem.Person.Name

	case RefTypeSurname:
		uri = f.Uri
		target = f.Node.Name

	case RefTypeOrigin:
		origin := mem.Origin
		uri = origin.Family.Uri
		target = origin.Person.Name
	}

	if target == nil || target == source {
		return
	}

	return proto.Location{
		URI:   uri,
		Range: TokenToRange(target),
	}, nil
}

func getDefinition(uri Uri, pos Position) (ref *Ref, err error) {
	err = root.UpdateDirty()

	if err != nil {
		return
	}

	uri = NormalizeUri(uri)

	ref = root.GetRefByPosition(uri, pos)

	return
}
