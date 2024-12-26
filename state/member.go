package state

import (
	"iter"
	"slices"

	. "github.com/redexp/familymarkup-lsp/types"
)

type Member struct {
	Name    string
	Aliases []string
	Node    *Node
	Refs    Refs
	InfoUri Uri
	Family  *Family
	Origin  *Member
}

func (member *Member) GetUniqName() string {
	family := member.Family

	for name := range member.NamesIter() {
		_, exist := family.Duplicates[name]

		if !exist {
			return name
		}
	}

	return ""
}

func (member *Member) NamesIter() iter.Seq[string] {
	return func(yield func(string) bool) {
		if !yield(member.Name) {
			return
		}

		for _, name := range member.Aliases {
			if !yield(name) {
				break
			}
		}
	}
}

func (member *Member) HasName(name string) bool {
	return member.Name == name || slices.Contains(member.Aliases, name)
}
