package src

import (
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Completion(context *glsp.Context, params *proto.CompletionParams) (any, error) {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return nil, err
	}

	doc, err := tempDoc(uri)

	if err != nil {
		return nil, err
	}

	t, nodes, err := getTypeNode(doc, &params.Position)

	if err != nil {
		return nil, err
	}

	list := make([]proto.CompletionItem, 0)

	addAliases := func(aliases []string) {
		if aliases == nil {
			return
		}

		for _, value := range aliases {
			list = append(list, proto.CompletionItem{
				Kind:  pt(proto.CompletionItemKindVariable),
				Label: value,
			})
		}
	}

	addMembers := func(family *Family) {
		for _, member := range family.Members {
			list = append(list, proto.CompletionItem{
				Kind:  pt(proto.CompletionItemKindVariable),
				Label: member.Name,
			})

			addAliases(member.Aliases)
		}
	}

	root.UpdateDirty()

	if t == "surname-name" || t == "surname-nil" {
		family := root.FindFamily(toString(nodes[0], doc))

		if family != nil {
			addMembers(family)
			return list, nil
		}
	}

	onlyFamilies := t == "nil-name" || t == "surname"

	for family := range root.FamilyIter() {
		list = append(list, proto.CompletionItem{
			Kind:  pt(proto.CompletionItemKindVariable),
			Label: family.Name,
		})

		addAliases(family.Aliases)

		if onlyFamilies {
			continue
		}

		addMembers(family)
	}

	return list, nil
}
