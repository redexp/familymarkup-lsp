package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"

	proto "github.com/tliron/glsp/protocol_3_16"
)

func Definition(_ *Ctx, params *proto.DefinitionParams) (res any, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	famMem, err := getDefinition(uri, params.Position)

	if err != nil || famMem == nil {
		return
	}

	family, member, source := famMem.Spread()

	var target *fm.Token

	if member != nil {
		target = member.Person.Name
		uri = member.Family.Uri
	} else if family != nil {
		target = family.Node.Name
		uri = family.Uri
	}

	if target == nil || target.IsEqual(source) {
		return
	}

	return proto.Location{
		URI:   uri,
		Range: TokenToRange(target),
	}, nil
}

func getDefinition(uri Uri, pos Position) (famMem *FamMem, err error) {
	uri, err = NormalizeUri(uri)

	if err != nil {
		return
	}

	err = root.UpdateDirty()

	if err != nil {
		return
	}

	// TODO: change famMem to Member.Origin
	famMem = root.GetFamMemByPosition(uri, pos)

	return
}
