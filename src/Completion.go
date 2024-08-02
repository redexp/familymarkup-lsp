package src

import (
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Completion(context *glsp.Context, params *proto.CompletionParams) (any, error) {
	doc, err := openDoc(params.TextDocument.URI)

	if err != nil {
		return nil, err
	}

	prev, target, next, err := doc.GetClosestHighlightCaptureByPosition(&params.Position)

	if err != nil {
		return nil, err
	}

	list := make([]proto.CompletionItem, 0)

	if target != nil && target.Node.Type() != "name" {
		return list, nil
	}

	caps := []*sitter.QueryCapture{prev, target, next}
	line := params.Position.Line

	for i, cap := range caps {
		if cap == nil {
			continue
		}

		node := cap.Node
		t := node.Type()

		if (t != "name" && t != "surname") || node.StartPoint().Row != line {
			caps[i] = nil
		}
	}

	prev = caps[0]
	next = caps[2]

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

	if prev != nil {
		value := prev.Node.Content([]byte(doc.Text))
		family := root.FindFamily(value)

		if family != nil {
			addMembers(family)
			return list, nil
		}
	}

	onlyFamilies := next != nil

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
