package state

import (
	"iter"

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
	NodeRefs   map[Uri]map[fm.Position]*Ref
	Duplicates map[string][]*Duplicate
	Listeners  map[string][]func()
)

func (ref *Ref) Spread() (*Family, *Member, *fm.Token) {
	return ref.Family, ref.Member, ref.Token
}

func (ref *Ref) TargetUri() Uri {
	switch ref.Type {
	case RefTypeNameSurname:
		return ref.Member.Family.Uri

	case RefTypeSurname:
		return ref.Family.Uri

	case RefTypeOrigin:
		return ref.Member.Origin.Family.Uri
	}

	return ref.Uri
}

func (nodes NodeRefs) RefsIter() iter.Seq2[*Ref, Uri] {
	return func(yield func(*Ref, Uri) bool) {
		for uri, refs := range nodes {
			for _, ref := range refs {
				if !yield(ref, uri) {
					return
				}
			}
		}
	}
}
