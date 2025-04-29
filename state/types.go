package state

import (
	. "github.com/redexp/familymarkup-lsp/types"
	fm "github.com/redexp/familymarkup-parser"
)

type Ref struct {
	Uri     Uri
	Surname *fm.Token
	Name    *fm.Token
	Person  *fm.Person
	Member  *Member
	Family  *Family
}

type Duplicate struct {
	Family *Family
	Member *Member
	Uri    string
}

type FamMem struct {
	Family *Family
	Member *Member
	Loc    fm.Loc
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
