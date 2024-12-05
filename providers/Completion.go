package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Completion(ctx *Ctx, params *proto.CompletionParams) (any, error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return nil, err
	}

	doc, err := TempDoc(uri)

	if err != nil {
		return nil, err
	}

	t, nodes, err := GetTypeNode(doc, &params.Position)

	if err != nil {
		return nil, err
	}

	list := make([]proto.CompletionItem, 0)
	hash := make(map[string]bool)

	if t == "surname" && len(nodes) == 1 {
		return list, nil
	}

	add := func(name string) {
		_, exist := hash[name]

		if exist {
			return
		}

		list = append(list, proto.CompletionItem{
			Kind:  P(proto.CompletionItemKindVariable),
			Label: name,
		})

		hash[name] = true
	}

	addAliases := func(aliases []string) {
		if aliases == nil {
			return
		}

		for _, alias := range aliases {
			add(alias)
		}
	}

	addMembers := func(family *Family) {
		for member := range family.MembersIter() {
			add(member.Name)
			addAliases(member.Aliases)
		}
	}

	root.UpdateDirty()

	if t == "nil-name" {
		family := root.FindFamily(ToString(nodes[0], doc))

		if family != nil {
			addMembers(family)
			return list, nil
		}
	}

	onlyFamilies := t == "surname-nil"

	for family := range root.FamilyIter() {
		add(family.Name)
		addAliases(family.Aliases)

		if onlyFamilies {
			continue
		}

		addMembers(family)
	}

	return list, nil
}
