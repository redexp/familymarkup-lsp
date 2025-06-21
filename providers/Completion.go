package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Completion(_ *Ctx, params *proto.CompletionParams) (res any, err error) {
	t, words, err := GetCompletionType(params.TextDocument.URI, params.Position)

	if err != nil || t == "" {
		return
	}

	// show names for surname
	if t == "surname" {
		t = "name"
	}

	list := make([]proto.CompletionItem, 0)
	hash := make(map[string]bool)

	for _, word := range words {
		hash[word] = true
	}

	kind := P(proto.CompletionItemKindVariable)

	add := func(names ...string) {
		for _, name := range names {
			_, exist := hash[name]

			if exist {
				continue
			}

			hash[name] = true

			list = append(list, proto.CompletionItem{
				Kind:  kind,
				Label: name,
			})
		}
	}

	addFamily := func(family *Family) {
		add(family.Name)
		add(family.Aliases...)
	}

	addMembers := func(family *Family) {
		for member := range family.MembersIter() {
			add(member.Name)
			add(member.Aliases...)
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
		surname := words[0]

		if len(words) > 1 {
			surname = words[1]
		}

		family := root.FindFamily(surname)

		if family != nil {
			addMembers(family)

			return list, nil
		}

		t = "name"
	}

	doc, err := GetDoc(params.TextDocument.URI)

	if err != nil {
		return
	}

	rel := doc.FindRelationByRange(PositionToRange(params.Position))

	if (t == "name |" || t == "name surname|") && rel != nil && rel.IsFamilyDef {
		t = "surname"
	}

	if t == "name |" || t == "name surname|" {
		name := words[0]

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
			if ref.Type == RefTypeSurname && ref.Token != nil {
				add(ref.Token.Text)
			}
		}
	}

	if t == "name" {
		for _, ref := range root.UnknownRefs {
			if ref.Person != nil {
				add(ref.Person.Name.Text)
			}
		}
	}

	return list, nil
}

// GetCompletionType
// "= |", []
// "name| surname", [string, string]
// "name |", [string]
// "| surname", [string]
// "= label|", [string]
// "name surname|", [string, string]
// "name" || "surname", [string]
// "", []
func GetCompletionType(uri Uri, pos Position) (t string, words []string, err error) {
	doc, err := GetDoc(uri)

	if err != nil {
		return
	}

	token := doc.GetTokenByPosition(pos)

	if token == nil {
		return
	}

	prev, next := doc.PrevNextNonSpaceTokens(token)

	mask := fm.TokenSpace | fm.TokenNewLine | fm.TokenEmptyLine
	blank := token.Type&mask == 0

	if blank && prev != nil && prev.SubType == fm.TokenEqual {
		return "= |", []string{}, nil
	}

	if blank && prev != nil && prev.Type == fm.TokenName && next != nil && next.Type == fm.TokenSurname {
		return "name| surname", []string{prev.Text, next.Text}, nil
	}

	if blank && prev != nil && prev.Type == fm.TokenName {
		return "name |", []string{prev.Text}, nil
	}

	if blank && next != nil && next.Type == fm.TokenSurname {
		return "| surname", []string{next.Text}, nil
	}

	if token.Type == fm.TokenWord && prev != nil && prev.SubType == fm.TokenEqual {
		return "= label|", []string{prev.Text}, nil
	}

	if token.Type == fm.TokenSurname && prev != nil && prev.Type == fm.TokenName {
		return "name surname|", []string{prev.Text, token.Text}, nil
	}

	if token.Type == fm.TokenName {
		return "name", []string{token.Text}, nil
	}

	if token.Type == fm.TokenSurname {
		return "surname", []string{token.Text}, nil
	}

	return "", []string{}, nil
}
