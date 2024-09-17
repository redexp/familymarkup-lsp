package state

import (
	"iter"

	. "github.com/redexp/familymarkup-lsp/types"
)

type Member struct {
	Id      string
	Name    string
	Aliases []string
	Node    *Node
	Refs    Refs
	Family  *Family
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
