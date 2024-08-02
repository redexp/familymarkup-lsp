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
}

type Members map[string]*Member

type Member struct {
	Id      string
	Name    string
	Aliases []string
}

func (root Families) Update(tree *Tree, text []byte) error {
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
				}
				root[value] = family

			case 1:
				family.Aliases = toAliases(cap.Node)

			case 2:
				member = &Member{
					Id:   value,
					Name: value,
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
	for _, family := range root {
		if strings.Contains(family.Name, name) {
			return family
		}
	}

	return nil
}

func createQuery(pattern string) (*sitter.Query, error) {
	return sitter.NewQuery([]byte(pattern), lang)
}
