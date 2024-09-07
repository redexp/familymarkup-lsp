package src

import (
	"iter"
	"math"
	"slices"

	familymarkup "github.com/redexp/tree-sitter-familymarkup"
)

var lang = familymarkup.GetLanguage()

type Root struct {
	Families    Families
	Duplicates  Duplicates
	NodeRefs    NodeRefs
	UnknownRefs []*Ref
	DirtyUris   UriSet
}

type Family struct {
	Id         string
	Name       string
	Aliases    []string
	Members    Members
	Duplicates Duplicates
	Refs       Refs
	Uri        Uri
	Node       *Node
	Root       *Root
}

type Member struct {
	Id      string
	Name    string
	Aliases []string
	Node    *Node
	Refs    Refs
	Family  *Family
}

type Ref struct {
	Uri     Uri
	Node    *Node
	Surname string
	Name    string
}

type Duplicate struct {
	Family *Family
	Member *Member
	Node   *Node
	Uri    string
}

type (
	Families   map[string]*Family
	Members    map[string]*Member
	NodeRefs   map[Uri]map[*Node]*Member
	UriSet     map[Uri]bool
	Duplicates map[string][]*Duplicate
	Refs       []*Ref
)

func createRoot() *Root {
	return &Root{
		Families:    make(Families),
		Duplicates:  make(Duplicates),
		NodeRefs:    make(NodeRefs),
		UnknownRefs: make([]*Ref, 0),
		DirtyUris:   make(UriSet),
	}
}

// root Methods

func (root *Root) Update(tree *Tree, text []byte, uri Uri) (err error) {
	q, err := createQuery(`
		(family_name 
			(name) @family-name
		)

		(name_ref
			(surname)
			(name)
		) @name_ref

		(sources
			(name) @sources-name
		)

		(targets
			(name_def
				(name) @name_def-name
			)
		)

		(name_def
			(new_surname
				(name) @name_def-new_surname
			)
		)
	`)

	if err != nil {
		return
	}

	defer q.Close()

	var family *Family

	for index, node := range queryIter(q, tree.RootNode()) {
		switch index {
		// new family
		case 0:
			family = root.AddFamily(uri, node, text)

		// name_ref
		case 1:
			root.AddRef(&Ref{
				Uri:     uri,
				Node:    node,
				Surname: node.NamedChild(0).Content(text),
				Name:    node.NamedChild(1).Content(text),
			})

		// sorces -> name
		case 2:
			name := node.Content(text)
			m := family.GetMember(name)

			if m != nil {
				root.AddRef(&Ref{
					Uri:     uri,
					Node:    node,
					Surname: family.Name,
					Name:    name,
				})
			} else {
				family.AddMember(node, text)
			}

		// new member or member ref
		case 3:
			rel := getClosestNode(node, "relation")

			if rel == nil {
				continue
			}

			arrow := rel.ChildByFieldName("arrow")

			if arrow != nil && arrow.Content(text) == "=" {
				family.AddMember(node, text)
			} else {
				root.AddRef(&Ref{
					Uri:     uri,
					Node:    node,
					Surname: family.Name,
					Name:    node.Content(text),
				})
			}

		// new_surname
		case 4:
			root.AddRef(&Ref{
				Uri:     uri,
				Node:    node,
				Surname: node.Content(text),
			})
		}
	}

	root.UpdateUnknownRefs()

	return
}

func (root *Root) UpdateUnknownRefs() {
	if len(root.UnknownRefs) == 0 {
		return
	}

	list := root.UnknownRefs
	root.UnknownRefs = make([]*Ref, 0)

	for _, ref := range list {
		root.AddRef(ref)
	}
}

func (root *Root) UpdateDirty() error {
	if len(root.DirtyUris) == 0 {
		return nil
	}

	uris := root.DirtyUris
	root.DirtyUris = UriSet{}
	root.UnknownRefs = filterRefs(root.UnknownRefs, uris)

	for uri := range root.NodeRefs {
		if uris.Has(uri) {
			delete(root.NodeRefs, uri)
		}
	}

	resetRefs := func(refs Refs) {
		for _, ref := range refs {
			if uris.Has(ref.Uri) {
				continue
			}

			root.UnknownRefs = append(root.UnknownRefs, ref)
		}
	}

	for family := range root.FamilyIter() {
		if uris.Has(family.Uri) {
			resetRefs(family.Refs)

			for _, member := range family.Members {
				resetRefs(member.Refs)
			}

			root.RemoveFamily(family)

			continue
		}

		for _, member := range family.Members {
			member.Refs = filterRefs(member.Refs, uris)
		}
	}

	for name, dups := range root.Duplicates {
		dups = slices.DeleteFunc(dups, func(dup *Duplicate) bool {
			return uris.Has(dup.Uri) || uris.Has(dup.Family.Uri)
		})

		if len(dups) == 0 {
			delete(root.Duplicates, name)
		} else {
			root.Duplicates[name] = dups
		}
	}

	tempDocs := make(Docs)

	for uri := range uris {
		if !docExist(uri) {
			continue
		}

		doc, err := tempDocs.Get(uri)

		if err != nil {
			return err
		}

		root.Update(doc.Tree, []byte(doc.Text), uri)
	}

	root.UpdateUnknownRefs()

	return nil
}

func (root *Root) AddFamily(uri Uri, node *Node, text []byte) *Family {
	name := node.Content(text)

	family := &Family{
		Id:         name,
		Name:       name,
		Aliases:    getAliases(node, text),
		Members:    make(Members),
		Duplicates: make(Duplicates),
		Refs:       make([]*Ref, 0),
		Uri:        uri,
		Node:       node,
		Root:       root,
	}

	names := append(family.Aliases, name)

	for _, name := range names {
		dup, exist := root.Families[name]

		if exist {
			addDuplicate(root.Duplicates, name, &Duplicate{
				Family: dup,
				Node:   node,
				Uri:    uri,
			})
		}

		root.Families[name] = family
	}

	return family
}

func (root *Root) FindFamily(name string) *Family {
	family, exist := root.Families[name]

	if exist {
		return family
	}

	source := []rune(name)
	min := uint(len(source))

	for key, f := range root.Families {
		diff := compareNames(source, []rune(key))

		if diff < min {
			min = diff
			family = f
		}
	}

	if min <= 2 {
		return family
	}

	return nil
}

func (root *Root) RemoveFamily(f *Family) {
	for key, v := range root.Families {
		if v == f {
			delete(root.Families, key)
		}
	}
}

func (root *Root) FamilyIter() iter.Seq[*Family] {
	return func(yield func(*Family) bool) {
		list := make(map[*Family]bool)

		for _, item := range root.Families {
			_, exist := list[item]

			if exist {
				continue
			}

			list[item] = true

			if !yield(item) {
				return
			}
		}
	}
}

func (root *Root) FindFamiliesByUri(uri Uri) []*Family {
	list := make([]*Family, 0)

	for family := range root.FamilyIter() {
		if family.Uri == uri {
			list = append(list, family)
		}
	}

	return list
}

func (root *Root) FindMember(surname string, name string) (family *Family, member *Member) {
	if surname == "" {
		return
	}

	family = root.FindFamily(surname)

	if family == nil || name == "" {
		return
	}

	member = family.GetMember(name)

	return
}

func (root *Root) AddRef(ref *Ref) {
	f, mem := root.FindMember(ref.Surname, ref.Name)

	if f != nil && ref.Name == "" {
		f.Refs = append(f.Refs, ref)
		return
	}

	if mem == nil {
		root.UnknownRefs = append(root.UnknownRefs, ref)
		return
	}

	mem.Refs = append(mem.Refs, ref)

	root.AddNodeRef(ref.Uri, nameRefName(ref.Node), mem)
}

func (root *Root) AddNodeRef(uri Uri, node *Node, mem *Member) {
	_, exist := root.NodeRefs[uri]

	if !exist {
		root.NodeRefs[uri] = make(map[*Node]*Member)
	}

	root.NodeRefs[uri][node] = mem
}

func (root *Root) GetMemberByUriNode(uri Uri, node *Node) *Member {
	_, exist := root.NodeRefs[uri]

	if !exist {
		return nil
	}

	return root.NodeRefs[uri][node]
}

// family Methods

func (family *Family) GetMember(name string) *Member {
	return family.Members[name]
}

func (family *Family) AddMember(node *Node, text []byte) {
	name := node.Content(text)
	aliases := getAliases(node, text)

	mem, exist := family.Members[name]

	if exist {
		addDuplicate(family.Duplicates, name, &Duplicate{
			Member: mem,
			Node:   node,
		})
	}

	member := &Member{
		Id:      name,
		Name:    name,
		Aliases: aliases,
		Node:    node,
		Refs:    make([]*Ref, 0),
		Family:  family,
	}

	family.Members[name] = member
	family.Root.AddNodeRef(family.Uri, node, member)

	aliasesNode := getAliasesNode(node)

	for i, alias := range aliases {
		mem, exist = family.Members[alias]

		if exist {
			addDuplicate(family.Duplicates, alias, &Duplicate{
				Member: mem,
				Node:   aliasesNode.NamedChild(i),
			})
		}

		family.Members[alias] = member
	}
}

func (family *Family) MembersIter() iter.Seq[*Member] {
	return func(yield func(*Member) bool) {
		list := make(map[*Member]bool)

		for _, item := range family.Members {
			_, exist := list[item]

			if exist {
				continue
			}

			list[item] = true

			if !yield(item) {
				return
			}
		}
	}
}

// member Methods

func (member *Member) GetUniqName() string {
	family := member.Family
	names := []string{member.Name}
	names = append(names, member.Aliases...)

	for _, name := range names {
		_, exist := family.Duplicates[name]

		if !exist {
			return name
		}
	}

	return ""
}

// UriSet Methods

func (uris UriSet) Set(uri Uri) {
	uris[uri] = true
}

func (uris UriSet) Has(uri Uri) bool {
	_, has := uris[uri]

	return has
}

func addDuplicate(duplicates Duplicates, name string, dup *Duplicate) {
	_, exist := duplicates[name]

	if !exist {
		duplicates[name] = make([]*Duplicate, 0)
	}

	duplicates[name] = append(duplicates[name], dup)
}

func filterRefs(refs []*Ref, uris UriSet) []*Ref {
	list := make([]*Ref, 0)

	for _, ref := range refs {
		if uris.Has(ref.Uri) {
			continue
		}

		list = append(list, ref)
	}

	return list
}

func getAliasesNode(node *Node) *Node {
	next := node.NextNamedSibling()

	if isNameAliases(next) {
		return next
	}

	parent := node.Parent()

	if isNameDef(parent) {
		return parent.ChildByFieldName("aliases")
	}

	return nil
}

func getAliases(nameNode *Node, text []byte) []string {
	node := getAliasesNode(nameNode)

	if node == nil {
		return make([]string, 0)
	}

	count := int(node.NamedChildCount())
	list := make([]string, count)

	for i := 0; i < count; i++ {
		list[i] = node.NamedChild(i).Content(text)
	}

	return list
}

func compareNames(a []rune, b []rune) uint {
	al := float64(len(a))
	bl := float64(len(b))
	max := uint(math.Max(al, bl))
	min := uint(math.Min(al, bl))
	diff := uint(max - min)

	if diff > 2 {
		return diff
	}

	for i := uint(0); i < min; i++ {
		if a[i] != b[i] {
			return max - 1 - i
		}
	}

	return diff
}
