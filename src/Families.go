package src

import (
	"strings"

	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
)

var lang = familymarkup.GetLanguage()

type Families map[string]*Family

type Family struct {
	Id      string
	Name    string
	Aliases []string
	Members Members
	Uri     Uri
	Node    *Node
}

type Members map[string]*Member

type Member struct {
	Id      string
	Name    string
	Aliases []string
	Node    *Node
}

func (root Families) Update(tree *Tree, text []byte, uri Uri) error {
	q, err := createQuery(`
		(family_name 
			(name) @family_name
			(name_aliases)? @family_aliases
		)

		(name_def
			(name) @member_name
			(name_aliases)? @member_aliases
		)
	`)

	if err != nil {
		return err
	}

	uri = toUri(uri)

	c := sitter.NewQueryCursor()
	c.Exec(q, tree.RootNode())
	defer c.Close()

	var family *Family
	var member *Member

	toAliases := func(node *Node) []string {
		count := int(node.NamedChildCount())
		list := make([]string, count)

		for i := 0; i < count; i++ {
			list[i] = node.NamedChild(i).Content(text)
		}

		return list
	}

	for {
		match, ok := c.NextMatch()

		if !ok {
			break
		}

		for _, cap := range match.Captures {
			node := cap.Node
			value := node.Content(text)

			switch cap.Index {
			case 0:
				family = &Family{
					Id:      value,
					Name:    value,
					Members: Members{},
					Uri:     uri,
					Node:    node,
				}
				root[value] = family

			case 1:
				family.Aliases = toAliases(cap.Node)

			case 2:
				member = &Member{
					Id:   value,
					Name: value,
					Node: node,
				}
				family.Members[value] = member

			case 3:
				member.Aliases = toAliases(cap.Node)
			}
		}
	}

	return nil
}

func (root Families) FindFamily(name string) *Family {
	var found *Family
	var foundAlias *Family

	for _, item := range root {
		n, a := compareNameAliases(item.Name, item.Aliases, name)

		if n == 1 || a == 1 {
			return item
		}

		if n == 2 {
			found = item
		}

		if found == nil && n == 3 {
			found = item
		}

		if a == 2 {
			foundAlias = item
		}

		if foundAlias == nil && a == 3 {
			foundAlias = item
		}
	}

	if found != nil {
		return found
	}

	return foundAlias
}

func (root Families) FindFamilyByNode(doc *TextDocument, node *Node) *Family {
	return root.FindFamily(node.Content([]byte(doc.Text)))
}

func (root Families) FindFamiliesByUri(uri Uri) []*Family {
	list := make([]*Family, 0)

	for _, family := range root {
		if family.Uri == uri {
			list = append(list, family)
		}
	}

	return list
}

func (family *Family) FindMember(name string) *Member {
	var found *Member
	var foundAlias *Member

	for _, item := range family.Members {
		n, a := compareNameAliases(item.Name, item.Aliases, name)

		if n == 1 || a == 1 {
			return item
		}

		if n == 2 {
			found = item
		}

		if found == nil && n == 3 {
			found = item
		}

		if a == 2 {
			foundAlias = item
		}

		if foundAlias == nil && a == 3 {
			foundAlias = item
		}
	}

	if found != nil {
		return found
	}

	return foundAlias
}

func (family *Family) FindMemberByNode(doc *TextDocument, node *Node) *Member {
	return family.FindMember(node.Content([]byte(doc.Text)))
}

func createQuery(pattern string) (*sitter.Query, error) {
	return sitter.NewQuery([]byte(pattern), lang)
}

func compareNameAliases(name string, aliases []string, value string) (uint8, uint8) {
	a := uint8(0)

	if name == value {
		return 1, 0
	}

	for _, alias := range aliases {
		if alias == value {
			return 0, 1
		}

		if strings.HasPrefix(alias, value) {
			a = 2
		}

		if a == 0 && strings.Contains(alias, value) {
			a = 3
		}
	}

	if strings.HasPrefix(name, value) {
		return 2, a
	}

	if strings.Contains(name, value) {
		return 3, a
	}

	return 0, a
}
