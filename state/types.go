package state

import (
	. "github.com/redexp/familymarkup-lsp/types"
)

type Ref struct {
	Uri     Uri
	Node    *Node
	Surname string
	Name    string
	Member  *Member
}

type Duplicate struct {
	Family *Family
	Member *Member
	Node   *Node
	Uri    string
}

type FamMem struct {
	Family *Family
	Member *Member
	Node   *Node
}

type (
	Families   map[string]*Family
	Members    map[string]*Member
	NodeRefs   map[Uri]map[string]*FamMem
	Duplicates map[string][]*Duplicate
	Refs       []*Ref
	Listeners  map[string][]func()
)
