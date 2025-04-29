package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Completion(ctx *Ctx, params *proto.CompletionParams) (res any, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	err = root.UpdateDirty()

	if err != nil {
		return
	}

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

// GetTypeNode
// "= |", []
// "= label|", [Loc]
// "name" || "surname", [Loc]
// "name surname|", [Loc, Loc]
// "name |", [Loc]
// "name| surname", [Loc, Loc]
// "| surname", [Loc]
// "nil", []
func GetTypeNode(doc *TextDocument, pos *Position) (t string, nodes []*Node, err error) {
	prev, target, next, err := doc.GetClosestHighlightCaptureByPosition(pos)

	if err != nil {
		return
	}

	caps := []*QueryCapture{prev, target, next}
	nodes = make([]*Node, 3)
	line := uint(pos.Line)

	for i, cap := range caps {
		if cap == nil || cap.Node.StartPosition().Row != line {
			continue
		}

		nodes[i] = &cap.Node
	}

	if nodes[0] != nil && nodes[0].Kind() == "eq" {
		if nodes[1] == nil && nodes[0].StartPosition().Row == uint(pos.Line) {
			return "= |", []*Node{}, nil
		}

		if nodes[1] != nil && nodes[1].Kind() == "words" {
			return "= label|", []*Node{nodes[1]}, nil
		}
	}

	for i, node := range nodes {
		if node == nil {
			continue
		}

		nt := node.Kind()

		if nt != "name" && nt != "surname" {
			nodes[i] = nil
			continue
		}

		parent := node.Parent()
		parentType := ""
		if parent != nil {
			parentType = parent.Kind()
		}

		if parentType == "name_aliases" {
			if i != 1 {
				nodes[i] = nil
				continue
			}

			return nt, []*Node{node}, nil
		}
	}

	if nodes[0] != nil {
		if nodes[1] != nil {
			return "name surname|", nodes[0:2], nil
		}

		return "name |", nodes[0:1], nil
	}

	node := nodes[1]

	if node != nil {
		if nodes[2] != nil {
			return "name| surname", nodes[1:3], nil
		}

		t = node.Kind()
		p := node.Parent()
		nodes = []*Node{node}

		if p != nil && p.Kind() == "family_name" {
			t = "surname"
			return
		}

		return
	}

	if nodes[2] != nil {
		return "| surname", nodes[2:3], nil
	}

	return "nil", []*Node{}, nil
}
