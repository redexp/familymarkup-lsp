package src

import (
	"strings"

	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
)

var lang = familymarkup.GetLanguage()

type Root struct {
	Families    Families
	UnknownRefs []*Ref
	DirtyUris   UriSet
}

type Families map[string]*Family

type Family struct {
	Id      string
	Name    string
	Aliases []string
	Members Members
	Uri     Uri
	Node    *Node
}

type Members map[string]*Member

type Member struct {
	Id      string
	Name    string
	Aliases []string
	Node    *Node
	Refs    []*Ref
}

type Ref struct {
	Uri     Uri
	Node    *Node
	Surname string
	Name    string
}

type UriSet map[Uri]bool

func createRoot() *Root {
	return &Root{
		Families:    Families{},
		UnknownRefs: make([]*Ref, 0),
		DirtyUris:   UriSet{},
	}
}

func (root *Root) Update(tree *Tree, text []byte, uri Uri) error {
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
	defer q.Close()

	if err != nil {
		return err
	}

	uri = toUri(uri)

	c := createCursor(q, tree)
	defer c.Close()

	var family *Family

	getAliases := func(nameNode *Node) []string {
		node := nameNode.NextNamedSibling()

		if node == nil || node.Type() != "name_aliases" {
			return make([]string, 0)
		}

		count := int(node.NamedChildCount())
		list := make([]string, count)

		for i := 0; i < count; i++ {
			list[i] = node.NamedChild(i).Content(text)
		}

		return list
	}

	addMember := func(node *Node, value string) {
		family.Members[value] = &Member{
			Id:      value,
			Name:    value,
			Aliases: getAliases(node),
			Node:    node,
			Refs:    make([]*Ref, 0),
		}
	}

	addRef := func(member *Member, node *Node) {
		member.Refs = append(member.Refs, &Ref{
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
			value := node.Content(text)

			switch cap.Index {
			// new family
			case 0:
				family = &Family{
					Id:      value,
					Name:    value,
					Aliases: getAliases(node),
					Members: Members{},
					Uri:     uri,
					Node:    node,
				}
				root.Families[value] = family

			// name_ref
			case 1:
				addRefByNode(uri, node, text)

			// member ref in current family or new member
			case 2:
				m := family.FindMember(value)

				if m != nil {
					addRef(m, node)
				} else {
					addMember(node, value)
				}

			// new member or member ref
			case 3:
				rel := getClosestNode(node, "relation")

				if rel == nil {
					continue
				}

				arrow := rel.ChildByFieldName("arrow")

				if arrow != nil && arrow.Content(text) == "=" {
					addMember(node, value)
				} else {
					m := family.FindMember(value)

					if m != nil {
						addRef(m, node)
					} else {
						root.UnknownRefs = append(root.UnknownRefs, &Ref{
							Uri:     uri,
							Node:    node,
							Surname: family.Name,
							Name:    value,
						})
					}
				}
			}
		}
	}

	root.UpdateUnknownRefs()

	return nil
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
			m.Refs = append(m.Refs, &Ref{
				Uri:  ref.Uri,
				Node: ref.Node,
			})
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
	defer q.Close()

	if err != nil {
		return err
	}

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
				addRefByNode(uri, cap.Node, text)
			}
		}

		c.Close()
	}

	root.UpdateUnknownRefs()

	return nil
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

	member = family.FindMember(name)

	return
}

func (family *Family) FindMember(name string) *Member {
	found, exist := family.Members[name]

	if exist {
		return found
	}

	var foundAlias *Member

	for _, item := range family.Members {
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

func addRefByNode(uri Uri, node *Node, text []byte) {
	surname := node.NamedChild(0).Content(text)
	name := node.NamedChild(1).Content(text)

	_, m := root.FindMember(surname, name)

	if m != nil {
		m.Refs = append(m.Refs, &Ref{
			Uri:  uri,
			Node: node,
		})
	} else {
		root.UnknownRefs = append(root.UnknownRefs, &Ref{
			Uri:     uri,
			Node:    node,
			Surname: surname,
			Name:    name,
		})
	}
}
