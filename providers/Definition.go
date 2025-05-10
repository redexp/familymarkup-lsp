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

	family, member, source, err := getDefinition(uri, params.Position)

	if err != nil {
		return
	}

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

func getDefinition(uri Uri, pos Position) (family *Family, member *Member, token *fm.Token, err error) {
	uri, err = NormalizeUri(uri)

	if err != nil {
		return
	}

	err = root.UpdateDirty()

	if err != nil {
		LogDebug("getDefinition UpdateDirty %s", err)
	}

	famMem := root.GetFamMemByPosition(uri, pos)

	if famMem == nil {
		return
	}

	family = famMem.Family
	member = famMem.Member
	token = famMem.Token

	return
}
