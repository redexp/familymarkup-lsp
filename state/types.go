package state

import (
	. "github.com/redexp/familymarkup-lsp/types"
	fm "github.com/redexp/familymarkup-parser"
)

type RefType int

const (
	RefTypeSurname RefType = 1 + iota
	RefTypeName
	RefTypeNameSurname
	RefTypeOrigin
)

type Ref struct {
	Type    RefType
	Uri     Uri
	Surname *fm.Token
	Person  *fm.Person
	Family  *Family
	Member  *Member
}

type Duplicate struct {
	Family *Family
	Member *Member
	Uri    string
}

type FamMem struct {
	Family *Family
	Member *Member
	Person *fm.Person
	Token  *fm.Token
}

type (
	Families   map[string]*Family
	Members    map[string]*Member
	NodeRefs   map[Uri]map[string]*FamMem
	Duplicates map[string][]*Duplicate
	Refs       []*Ref
	Listeners  map[string][]func()
)

func (famMem *FamMem) Spread() (*Family, *Member, *fm.Token) {
	return famMem.Family, famMem.Member, famMem.Token
}
