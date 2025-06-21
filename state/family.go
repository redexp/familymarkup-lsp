package state

import (
	fm "github.com/redexp/familymarkup-parser"
	"iter"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
)

type Family struct {
	Name       string
	Aliases    []string
	Members    Members
	Duplicates Duplicates
	Uri        Uri
	Node       *fm.Family
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

func (family *Family) FindMemberByPerson(person *fm.Person) *Member {
	for member := range family.MembersIter() {
		if member.Person == person {
			return member
		}
	}

	return nil
}

func (family *Family) AddMember(person *fm.Person) *Member {
	name := person.Name.Text
	aliases := TokensToStrings(person.Aliases)

	surname := ""

	if person.Surname != nil {
		surname = person.Surname.Text
	}

	mem, exist := family.Members[name]

	if exist {
		family.AddDuplicate(name, mem)
	}

	member := &Member{
		Name:    name,
		Aliases: aliases,
		Surname: surname,
		Person:  person,
		Family:  family,
	}

	family.Members[name] = member

	for _, alias := range aliases {
		mem, exist = family.Members[alias]

		if exist {
			family.AddDuplicate(alias, mem)
		}

		family.Members[alias] = member
	}

	return member
}

func (family *Family) AddDuplicate(name string, member *Member) {
	addDuplicate(family.Duplicates, name, &Duplicate{
		Member: member,
	})
}

func (family *Family) MembersIter() iter.Seq[*Member] {
	return func(yield func(*Member) bool) {
		check := createUniqYield(yield)

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

func (family *Family) GetRefsIter() iter.Seq2[*Ref, Uri] {
	return func(yield func(*Ref, Uri) bool) {
		for uri, refs := range family.Root.NodeRefs {
			for _, ref := range refs {
				if ref.Family != family {
					continue
				}

				if !yield(ref, uri) {
					return
				}
			}
		}
	}
}
