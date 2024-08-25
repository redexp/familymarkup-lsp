package src

import (
	"strings"

	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
)

var lang = familymarkup.GetLanguage()

type Root struct {
	Families    Families
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
	Uri        Uri
	Node       *Node
	Root       *Root
}

type Member struct {
	Id      string
	Name    string
	Aliases []string
	Node    *Node
	Refs    []*Ref
	Family  *Family
}

type Ref struct {
	Uri     Uri
	Node    *Node
	Surname string
	Name    string
}

type Duplicate struct {
	Member *Member
	Node   *Node
}

type (
	Families   map[string]*Family
	Members    map[string]*Member
	NodeRefs   map[Uri]map[*Node]*Member
	UriSet     map[Uri]bool
	Duplicates map[string][]*Duplicate
)

func createRoot() *Root {
	return &Root{
		Families:    make(Families),
		NodeRefs:    make(NodeRefs),
		UnknownRefs: make([]*Ref, 0),
		DirtyUris:   make(UriSet),
	}
}

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

(relation
	(targets
		(name_def
			(name) @name_def-name
		)
	)
)
	`)

	if err != nil {
		return
	}

	defer q.Close()

	c := createCursor(q, tree)
	defer c.Close()

	var family *Family

	addRef := func(member *Member, node *Node) {
		root.AddMemberRef(member, &Ref{
			Uri:  uri,
			Node: node,
		})
	}

	for {
		match, ok := c.NextMatch()

		if !ok {
			break
		}

		for _, cap := range match.Captures {
			node := cap.Node

			switch cap.Index {
			// new family
			case 0:
				family = root.AddFamily(uri, node, text)

			// name_ref
			case 1:
				root.addRefByNode(uri, node, text)

			// sorces -> name
			case 2:
				name := node.Content(text)
				m := family.GetMember(name)

				if m != nil {
					addRef(m, node)
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
					name := node.Content(text)
					m := family.GetMember(name)

					if m != nil {
						addRef(m, node)
					} else {
						root.UnknownRefs = append(root.UnknownRefs, &Ref{
							Uri:     uri,
							Node:    node,
							Surname: family.Name,
							Name:    name,
						})
					}
				}
			}
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
		_, m := root.FindMember(ref.Surname, ref.Name)

		if m != nil {
			root.AddMemberRef(m, ref)
		} else {
			root.UnknownRefs = append(root.UnknownRefs, ref)
		}
	}
}

func (root *Root) UpdateDirty() error {
	if len(root.DirtyUris) == 0 {
		return nil
	}

	uris := root.DirtyUris
	root.DirtyUris = UriSet{}

	refsUris := UriSet{}

	for uri := range root.NodeRefs {
		if uris.Has(uri) {
			delete(root.NodeRefs, uri)
		}
	}

	for id, family := range root.Families {
		if uris.Has(family.Uri) {
			for _, member := range family.Members {
				for _, ref := range member.Refs {
					if uris.Has(ref.Uri) {
						continue
					}

					refsUris.Set(ref.Uri)
				}
			}

			delete(root.Families, id)

			continue
		}

		for _, member := range family.Members {
			member.Refs = filterRefs(member.Refs, uris)
		}
	}

	root.UnknownRefs = filterRefs(root.UnknownRefs, uris)

	for uri := range uris {
		doc, err := openDoc(uri)

		if err != nil {
			return err
		}

		root.Update(doc.Tree, []byte(doc.Text), uri)
	}

	q, err := createQuery(`
		(name_ref
			(surname)
			(name)
		) @name_ref
	`)

	if err != nil {
		return err
	}

	defer q.Close()

	for uri := range refsUris {
		doc, err := openDoc(uri)

		if err != nil {
			return err
		}

		c := createCursor(q, doc.Tree)
		text := []byte(doc.Text)

		for {
			match, ok := c.NextMatch()

			if !ok {
				break
			}

			for _, cap := range match.Captures {
				node := cap.Node
				surname := node.NamedChild(0).Content(text)
				name := node.NamedChild(1).Content(text)

				f, m := root.FindMember(surname, name)

				if f == nil || !uris.Has(f.Uri) {
					continue
				}

				root.AddMemberRef(m, &Ref{
					Uri:     uri,
					Node:    node,
					Surname: surname,
					Name:    name,
				})
			}
		}

		c.Close()
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
		Members:    Members{},
		Duplicates: make(Duplicates),
		Uri:        uri,
		Node:       node,
		Root:       root,
	}

	root.Families[name] = family

	return family
}

func (root *Root) FindFamily(name string) *Family {
	found, exist := root.Families[name]

	if exist {
		return found
	}

	var foundAlias *Family

	for _, item := range root.Families {
		n, a := compareNameAliases(item.Name, item.Aliases, name)

		if n == 1 || a == 1 {
			return item
		}

		if n == 2 {
			found = item
		}

		if found == nil && n == 3 {
			found = item
		}

		if a == 2 {
			foundAlias = item
		}

		if foundAlias == nil && a == 3 {
			foundAlias = item
		}
	}

	if found != nil {
		return found
	}

	return foundAlias
}

func (root *Root) FindFamiliesByUri(uri Uri) []*Family {
	list := make([]*Family, 0)

	for _, family := range root.Families {
		if family.Uri == uri {
			list = append(list, family)
		}
	}

	return list
}

func (root *Root) FindMember(surname string, name string) (family *Family, member *Member) {
	family = root.FindFamily(surname)

	if family == nil {
		return
	}

	member = family.GetMember(name)

	return
}

func (family *Family) GetMember(name string) *Member {
	return family.Members[name]
}

func (root *Root) GetMemberByUriNode(uri Uri, node *Node) *Member {
	_, exist := root.NodeRefs[uri]

	if !exist {
		return nil
	}

	return root.NodeRefs[uri][node]
}

func (family *Family) AddMember(node *Node, text []byte) {
	name := node.Content(text)
	aliases := getAliases(node, text)

	mem, exist := family.Members[name]

	if exist {
		family.addDuplicate(name, mem, node)
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

	aliasesNode := getAliasesNode(node)

	for i, alias := range aliases {
		mem, exist = family.Members[alias]

		if exist {
			family.addDuplicate(alias, mem, aliasesNode.NamedChild(i))
		}

		family.Members[alias] = member
	}
}

func (family *Family) addDuplicate(name string, member *Member, node *Node) {
	_, exist := family.Duplicates[name]

	if !exist {
		family.Duplicates[name] = make([]*Duplicate, 0)
	}

	family.Duplicates[name] = append(family.Duplicates[name], &Duplicate{
		Member: member,
		Node:   node,
	})
}

func (uris UriSet) Set(uri Uri) {
	uris[uri] = true
}

func (uris UriSet) Has(uri Uri) bool {
	_, has := uris[uri]

	return has
}

func createQuery(pattern string) (*sitter.Query, error) {
	return sitter.NewQuery([]byte(pattern), lang)
}

func createCursor(q *sitter.Query, tree *sitter.Tree) *sitter.QueryCursor {
	c := sitter.NewQueryCursor()
	c.Exec(q, tree.RootNode())
	return c
}

func compareNameAliases(name string, aliases []string, value string) (uint8, uint8) {
	a := uint8(0)

	if name == value {
		return 1, 0
	}

	for _, alias := range aliases {
		if alias == value {
			return 0, 1
		}

		if strings.HasPrefix(alias, value) {
			a = 2
		}

		if a == 0 && strings.Contains(alias, value) {
			a = 3
		}
	}

	if strings.HasPrefix(name, value) {
		return 2, a
	}

	if strings.Contains(name, value) {
		return 3, a
	}

	return 0, a
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

func (root *Root) addRefByNode(uri Uri, node *Node, text []byte) {
	surname := node.NamedChild(0).Content(text)
	name := node.NamedChild(1).Content(text)

	_, m := root.FindMember(surname, name)

	root.AddMemberRef(m, &Ref{
		Uri:     uri,
		Node:    node,
		Surname: surname,
		Name:    name,
	})
}

func (root *Root) AddMemberRef(mem *Member, ref *Ref) {
	if mem == nil {
		root.UnknownRefs = append(root.UnknownRefs, ref)
		return
	}

	mem.Refs = append(mem.Refs, ref)

	uri := ref.Uri
	node := nameRefName(ref.Node)

	_, exist := root.NodeRefs[uri]

	if !exist {
		root.NodeRefs[uri] = make(map[*Node]*Member)
	}

	root.NodeRefs[uri][node] = mem
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
