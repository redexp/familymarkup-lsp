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
	Type   RefType
	Uri    Uri
	Person *fm.Person
	Family *Family
	Member *Member
	Token  *fm.Token
}

type Duplicate struct {
	Family *Family
	Member *Member
	Uri    string
}

type (
	Families   map[string]*Family
	Members    map[string]*Member
	NodeRefs   map[Uri]map[string]*Ref
	Duplicates map[string][]*Duplicate
	Refs       []*Ref
	Listeners  map[string][]func()
)

func (ref *Ref) Spread() (*Family, *Member, *fm.Token) {
	return ref.Family, ref.Member, ref.Token
}
