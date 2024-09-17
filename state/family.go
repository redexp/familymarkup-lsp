package state

import (
	"iter"

	. "github.com/redexp/familymarkup-lsp/types"
)

type Family struct {
	Id         string
	Name       string
	Aliases    []string
	Members    Members
	Duplicates Duplicates
	Refs       Refs
	Uri        Uri
	Node       *Node
	Root       *Root
}

func (family *Family) GetMember(name string) *Member {
	return family.Members[name]
}

func (family *Family) AddMember(node *Node, text []byte) {
	name := node.Content(text)
	aliases := getAliases(node, text)

	mem, exist := family.Members[name]

	if exist {
		addDuplicate(family.Duplicates, name, &Duplicate{
			Member: mem,
			Node:   node,
		})
	}

	member := &Member{
		Id:      name,
		Name:    name,
		Aliases: aliases,
		Node:    node,
		Refs:    make([]*Ref, 0),
		Family:  family,
	}

	family.Members[name] = member
	family.Root.AddNodeRef(family.Uri, node, member)

	aliasesNode := getAliasesNode(node)

	for i, alias := range aliases {
		mem, exist = family.Members[alias]

		if exist {
			addDuplicate(family.Duplicates, alias, &Duplicate{
				Member: mem,
				Node:   aliasesNode.NamedChild(i),
			})
		}

		family.Members[alias] = member
	}
}

func (family *Family) MembersIter() iter.Seq[*Member] {
	return func(yield func(*Member) bool) {
		list := make(map[*Member]bool)

		for _, item := range family.Members {
			_, exist := list[item]

			if exist {
				continue
			}

			list[item] = true

			if !yield(item) {
				return
			}
		}
	}
}

func (family *Family) NamesIter() iter.Seq[string] {
	return func(yield func(string) bool) {
		if !yield(family.Name) {
			return
		}

		for _, name := range family.Aliases {
			if !yield(name) {
				break
			}
		}
	}
}
