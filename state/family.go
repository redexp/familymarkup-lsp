package state

import (
	"iter"

	. "github.com/redexp/familymarkup-lsp/types"
)

type Family struct {
	Name       string
	Aliases    []string
	Members    Members
	Duplicates Duplicates
	Uri        Uri
	Node       *Node
	Root       *Root
}

func (family *Family) HasMember(name string) bool {
	_, exist := family.Members[name]

	return exist
}

func (family *Family) GetMember(name string) *Member {
	return family.Members[name]
}

func (family *Family) FindMember(name string) (mem *Member) {
	mem = family.GetMember(name)

	if mem != nil {
		return
	}

	source := []rune(name)
	min := uint(len(source))

	for key, m := range family.Members {
		diff := compareNames(source, []rune(key))

		if diff < min {
			min = diff
			mem = m
		}
	}

	if min <= 2 {
		return mem
	}

	return nil
}

func (family *Family) AddMember(node *Node, text []byte) *Member {
	name := node.Utf8Text(text)
	aliases := getAliases(node, text)

	surnameNode := node.Parent().ChildByFieldName("surname")
	surname := ""

	if surnameNode != nil {
		surname = surnameNode.Utf8Text(text)
	}

	return family.AddMemberName(node, name, aliases, surname)
}

func (family *Family) AddMemberName(node *Node, name string, aliases []string, surname string) *Member {
	mem, exist := family.Members[name]

	if exist {
		addDuplicate(family.Duplicates, name, &Duplicate{
			Member: mem,
		})
	}

	member := &Member{
		Name:    name,
		Aliases: aliases,
		Surname: surname,
		Node:    node,
		Refs:    make([]*Ref, 0),
		Family:  family,
	}

	family.Members[name] = member
	family.Root.AddNodeRef(family.Uri, &FamMem{Member: member, Node: node})

	for _, alias := range aliases {
		mem, exist = family.Members[alias]

		if exist {
			addDuplicate(family.Duplicates, alias, &Duplicate{
				Member: mem,
			})
		}

		family.Members[alias] = member
	}

	return member
}

func (family *Family) MembersIter() iter.Seq[*Member] {
	return func(yield func(*Member) bool) {
		check := createIterCheck(yield)

		for _, item := range family.Members {
			if check(item) {
				return
			}
		}

		for _, dups := range family.Duplicates {
			for _, dup := range dups {
				if check(dup.Member) {
					return
				}
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
