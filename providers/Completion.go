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

	root.UpdateDirty()

	doc, err := TempDoc(uri)

	if err != nil {
		return nil, err
	}

	t, nodes, err := GetTypeNode(doc, &params.Position)

	if err != nil || t == "nil" {
		return nil, err
	}

	// show names for surname
	if t == "surname" {
		t = "name"
	}

	list := make([]proto.CompletionItem, 0)
	hash := make(map[string]bool)

	for _, node := range nodes {
		hash[ToString(node, doc)] = true
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

	addFamily := func(family *Family) {
		add(family.Name)
		addAliases(family.Aliases)
	}

	addMembers := func(family *Family) {
		for member := range family.MembersIter() {
			add(member.Name)
			addAliases(member.Aliases)
		}
	}

	if t == "= |" || t == "= label|" {
		for _, labels := range root.Labels {
			for _, label := range labels {
				add(label)
			}
		}

		return list, nil
	}

	if t == "| surname" || t == "name| surname" {
		surname := nodes[0]

		if len(nodes) > 1 {
			surname = nodes[1]
		}

		family := root.FindFamily(ToString(surname, doc))

		if family != nil {
			addMembers(family)

			return list, nil
		}

		t = "name"
	}

	if (t == "name |" || t == "name surname|") && IsNameDef(nodes[0].Parent()) {
		t = "surname"
	}

	if t == "name |" || t == "name surname|" {
		name := ToString(nodes[0], doc)

		for member := range root.MembersIter() {
			if member.HasName(name) {
				addFamily(member.Family)
			}
		}

		if len(list) > 0 {
			return list, nil
		}

		t = "surname"
	}

	for family := range root.FamilyIter() {
		if t == "surname" {
			addFamily(family)
		} else {
			addMembers(family)
		}
	}

	if t == "surname" {
		for _, ref := range root.UnknownRefs {
			if ref.Surname != "" {
				add(ref.Surname)
			}
		}
	}

	if t == "name" {
		for _, ref := range root.UnknownRefs {
			if ref.Name != "" {
				add(ref.Name)
			}
		}
	}

	return list, nil
}
