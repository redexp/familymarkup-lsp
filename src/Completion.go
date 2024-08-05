package src

import (
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Completion(context *glsp.Context, params *proto.CompletionParams) (any, error) {
	doc, err := openDoc(params.TextDocument.URI)

	if err != nil {
		return nil, err
	}

	t, nodes, err := getTypeNode(doc, &params.Position)

	if err != nil {
		return nil, err
	}

	list := make([]proto.CompletionItem, 0)

	kind := proto.CompletionItemKindVariable

	addAliases := func(aliases []string) {
		if aliases == nil {
			return
		}

		for _, value := range aliases {
			list = append(list, proto.CompletionItem{
				Kind:  &kind,
				Label: value,
			})
		}
	}

	addMembers := func(family *Family) {
		for _, member := range family.Members {
			list = append(list, proto.CompletionItem{
				Kind:  &kind,
				Label: member.Name,
			})

			addAliases(member.Aliases)
		}
	}

	if t == "surname-name" || t == "surname-nil" {
		family := root.FindFamily(toString(nodes[0], doc))

		if family != nil {
			addMembers(family)
			return list, nil
		}
	}

	onlyFamilies := t == "nil-name" || t == "surname"

	for _, family := range root {
		list = append(list, proto.CompletionItem{
			Kind:  &kind,
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
