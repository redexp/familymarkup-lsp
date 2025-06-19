package state

import (
	fm "github.com/redexp/familymarkup-parser"
	"iter"
	"slices"

	. "github.com/redexp/familymarkup-lsp/types"
)

type Member struct {
	Person  *fm.Person
	Name    string
	Aliases []string
	Surname string
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

func (member *Member) NormalizeName(name string) (res string) {
	if member.HasName(name) {
		return name
	}

	runeName := []rune(name)
	min := uint(len(runeName))

	for n := range member.NamesIter() {
		diff := compareNames([]rune(n), runeName)

		if diff <= 2 && diff < min {
			min = diff
			res = n
		}
	}

	return
}

func (member *Member) GetRefsIter() iter.Seq2[Uri, *fm.Person] {
	return func(yield func(Uri, *fm.Person) bool) {
		for uri, refs := range member.Family.Root.NodeRefs {
			for _, famMem := range refs {
				if famMem.Member != member {
					continue
				}

				if !yield(uri, famMem.Person) {
					return
				}
			}
		}
	}
}

func (member *Member) HasRef() bool {
	for range member.GetRefsIter() {
		return true
	}

	return false
}
