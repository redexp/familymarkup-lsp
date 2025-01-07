package state

import (
	"fmt"
	"iter"
	"slices"
	"strings"
	"sync"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
)

type Root struct {
	Folders      UriSet
	Families     Families
	Duplicates   Duplicates
	NodeRefs     NodeRefs
	UnknownRefs  []*Ref
	UnknownFiles Files
	DirtyUris    UriSet
	Listeners    Listeners
	Log          func(string, ...any)

	UpdateLock sync.Mutex
}

func CreateRoot(logger func(string, ...any)) *Root {
	return &Root{
		Folders:      make(UriSet),
		Families:     make(Families),
		Duplicates:   make(Duplicates),
		NodeRefs:     make(NodeRefs),
		UnknownRefs:  make([]*Ref, 0),
		UnknownFiles: make(Files),
		DirtyUris:    make(UriSet),
		Listeners:    make(Listeners),
		Log:          logger,
	}
}

func (root *Root) SetFolders(folders []Uri) (err error) {
	root.Folders = make(UriSet)

	for _, uri := range folders {
		uri, err = NormalizeUri(uri)

		if err != nil {
			return
		}

		root.Folders.Set(uri)
	}

	type TextTree struct {
		Text []byte
		Tree *Tree
		Uri  Uri
		MD   bool
	}

	textTrees := make(chan TextTree, 3)

	go func() {
		for uri := range root.Folders {
			WalkFiles(uri, AllExt, func(uri Uri, ext string) error {
				if slices.Contains(MarkdownExt, ext) {
					textTrees <- TextTree{
						Uri: uri,
						MD:  true,
					}
					return nil
				}

				tree, text, err := GetTreeText(uri)

				if err != nil {
					return err
				}

				textTrees <- TextTree{
					Text: text,
					Tree: tree,
					Uri:  uri,
				}

				return nil
			})
		}

		close(textTrees)
	}()

	for item := range textTrees {
		if item.MD {
			root.AddUnknownFile(item.Uri)
			continue
		}

		err = root.Update(item.Tree, item.Text, item.Uri)

		if err != nil {
			return
		}
	}

	root.UpdateUnknownRefs()
	root.UpdateUnknownFiles()

	return
}

func (root *Root) Update(tree *Tree, text []byte, uri Uri) (err error) {
	q, err := CreateQuery(`
		(family_name 
			(name) @family-name
		)

		(name_ref) @name_ref

		(sources
			(name) @sources-name
		)

		(targets
			(name_def
				(name) @name_def-name
			)
		)

		(name_def
			(surname) @name_def-surname
		)
	`)

	if err != nil {
		return
	}

	defer q.Close()

	var family *Family

	for index, node := range QueryIter(q, tree.RootNode(), text) {
		if node.IsMissing() {
			continue
		}

		switch index {
		// new family
		case 0:
			family = root.AddFamily(uri, node, text)

		// name_ref
		case 1:
			name, surname := GetNameSurname(node)

			root.AddRef(&Ref{
				Uri:     uri,
				Node:    node,
				Surname: surname.Utf8Text(text),
				Name:    name.Utf8Text(text),
			})

		// sorces -> name
		case 2:
			name := node.Utf8Text(text)

			if !IsFamilyRelation(node) || family.HasMember(name) {
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
			if !IsFamilyRelation(node) {
				root.AddRef(&Ref{
					Uri:     uri,
					Node:    node,
					Surname: family.Name,
					Name:    node.Utf8Text(text),
				})

				continue
			}

			mem := family.AddMember(node, text)

			surname := node.Parent().ChildByFieldName("surname")

			if surname == nil {
				continue
			}

			// post update check for surname and ref from that surname to this member
			root.AddUnknownRef(&Ref{
				Uri:     uri,
				Node:    surname,
				Surname: surname.Utf8Text(text),
				Member:  mem,
			})

		// new surname
		case 4:
			root.AddRef(&Ref{
				Uri:     uri,
				Node:    node,
				Surname: node.Utf8Text(text),
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
	root.UnknownRefs = Refs{}

	for _, ref := range list {
		if ref.Member == nil {
			continue
		}

		f := root.FindFamily(ref.Surname)

		if f == nil {
			continue
		}

		origin := ref.Member
		var targetNode *Node

		// find first ref of this member in that file (usualy as mether in family relation)
		for _, ref := range origin.Refs {
			if ref.Uri == f.Uri {
				targetNode = ref.Node
				break
			}
		}

		if targetNode == nil {
			root.AddUnknownRef(ref)
			continue
		}

		mem := f.AddMemberName(targetNode, origin.Name, origin.Aliases)
		mem.Origin = ref.Member
	}

	for _, ref := range list {
		if ref.Member == nil {
			root.AddRef(ref)
		}
	}
}

func (root *Root) UpdateUnknownFiles() {
	files := root.UnknownFiles

	if len(files) == 0 {
		return
	}

	tree := &FileTree{
		Children: make(FilesTree),
	}

	for uri, file := range files {
		item := tree
		var family *Family
		var member *Member

		for _, name := range file.Path {
			next, exist := item.Children[name]

			if exist {
				if next.Family != nil {
					family = next.Family
				}

				item = next
				continue
			}

			item.Children[name] = &FileTree{
				Children: make(FilesTree),
			}

			item = item.Children[name]

			if family == nil {
				family = root.FindFamily(name)
				item.Family = family
			} else {
				member = family.GetMember(name)

				if member != nil {
					member.InfoUri = file.Uri
					delete(root.UnknownFiles, uri)
					break
				}
			}
		}
	}
}

func (root *Root) UpdateDirty() error {
	root.UpdateLock.Lock()
	defer root.UpdateLock.Unlock()

	if len(root.DirtyUris) == 0 {
		return nil
	}

	uris := root.DirtyUris
	root.DirtyUris = UriSet{}
	root.UnknownRefs = filterRefs(root.UnknownRefs, uris)

	for uri, state := range uris {
		if !IsMarkdownUri(uri) {
			continue
		}

		uris.Remove(uri)

		_, exist := root.UnknownFiles[uri]

		deleted := state == FileDelete

		if exist {
			if deleted {
				delete(root.UnknownFiles, uri)
			}

			continue
		}

		if deleted {
			for mem := range root.MembersIter() {
				if mem.InfoUri == uri {
					mem.InfoUri = ""
					break
				}
			}
		} else {
			root.AddUnknownFile(uri)
		}
	}

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

			root.AddUnknownRef(ref)
		}
	}

	for family := range root.FamilyIter() {
		if uris.Has(family.Uri) {
			for member := range family.MembersIter() {
				resetRefs(member.Refs)

				if member.InfoUri != "" {
					root.AddUnknownFile(member.InfoUri)
				}
			}

			root.RemoveFamily(family)

			continue
		}

		for member := range family.MembersIter() {
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

	for uri, state := range uris {
		if state == FileDelete {
			continue
		}

		doc, err := tempDocs.Get(uri)

		if err != nil {
			return err
		}

		root.Update(doc.Tree, []byte(doc.Text), uri)
	}

	root.UpdateUnknownRefs()
	root.UpdateUnknownFiles()
	root.Trigger(RootOnUpdate)

	return nil
}

func (root *Root) AddFamily(uri Uri, node *Node, text []byte) *Family {
	name := node.Utf8Text(text)

	family := &Family{
		Name:       name,
		Aliases:    getAliases(node, text),
		Members:    make(Members),
		Duplicates: make(Duplicates),
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
				Uri:    uri,
			})
		}

		root.Families[name] = family
	}

	root.AddNodeRef(family.Uri, &FamMem{Family: family, Node: family.Node})

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
	for name, v := range root.Families {
		if v != f {
			continue
		}

		delete(root.Families, name)

		dups, exist := root.Duplicates[name]

		if !exist {
			continue
		}

		dups = slices.DeleteFunc(dups, func(d *Duplicate) bool {
			return d.Family == f
		})

		if len(dups) == 0 {
			delete(root.Duplicates, name)
		} else {
			root.Duplicates[name] = dups
		}
	}

	for uri, nodes := range root.NodeRefs {
		for key, item := range nodes {
			if item.Family == f && item.Member == nil {
				delete(nodes, key)

				root.AddUnknownRef(&Ref{
					Uri:     uri,
					Node:    item.Node,
					Surname: f.Name,
				})
			}
		}
	}
}

func (root *Root) FamilyIter() iter.Seq[*Family] {
	return func(yield func(*Family) bool) {
		check := createIterCheck(yield)

		for _, item := range root.Families {
			if check(item) {
				return
			}
		}

		for _, dups := range root.Duplicates {
			for _, dup := range dups {
				if check(dup.Family) {
					return
				}
			}
		}
	}
}

func (root *Root) MembersIter() iter.Seq[*Member] {
	return func(yield func(*Member) bool) {
		for f := range root.FamilyIter() {
			for mem := range f.MembersIter() {
				if !yield(mem) {
					return
				}
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

	member = family.FindMember(name)

	return
}

func (root *Root) AddRef(ref *Ref) {
	f, mem := root.FindMember(ref.Surname, ref.Name)

	if f != nil && ref.Name == "" {
		root.AddNodeRef(ref.Uri, &FamMem{Family: f, Node: ref.Node})
		return
	}

	if mem == nil {
		root.AddUnknownRef(ref)
		return
	}

	mem.Refs = append(mem.Refs, ref)

	if IsNameRef(ref.Node) {
		name, surname := GetNameSurname(ref.Node)
		root.AddNodeRef(ref.Uri, &FamMem{Member: mem, Node: name})
		root.AddNodeRef(ref.Uri, &FamMem{Family: f, Node: surname})
	} else {
		root.AddNodeRef(ref.Uri, &FamMem{Member: mem, Node: ref.Node})
	}
}

func (root *Root) AddNodeRef(uri Uri, famMem *FamMem) {
	_, exist := root.NodeRefs[uri]

	if !exist {
		root.NodeRefs[uri] = make(map[string]*FamMem)
	}

	pos := NodeToPosString(famMem.Node)

	_, exist = root.NodeRefs[uri][pos]

	if exist {
		return
	}

	if famMem.Member != nil && famMem.Member.Origin != nil {
		famMem = &FamMem{Member: famMem.Member.Origin}
	}

	root.NodeRefs[uri][pos] = famMem
}

func (root *Root) GetFamMem(uri Uri, node *Node) *FamMem {
	nodesMap, exist := root.NodeRefs[uri]

	if !exist {
		return nil
	}

	famMem, exist := nodesMap[NodeToPosString(node)]

	if !exist {
		return nil
	}

	return famMem
}

func (root *Root) GetFamilyByUriNode(uri Uri, node *Node) *Family {
	famMem := root.GetFamMem(uri, node)

	if famMem == nil {
		return nil
	}

	return famMem.Family
}

func (root *Root) GetMemberByUriNode(uri Uri, node *Node) *Member {
	famMem := root.GetFamMem(uri, node)

	if famMem == nil {
		return nil
	}

	return famMem.Member
}

func (root *Root) FindFolder(uri Uri) Uri {
	for folder := range root.Folders {
		if strings.HasPrefix(uri, folder) {
			return folder
		}
	}

	return ""
}

func (root *Root) AddUnknownRef(ref *Ref) {
	root.UnknownRefs = append(root.UnknownRefs, ref)
}

func (root *Root) AddUnknownFile(uri Uri) error {
	file, err := CreateFile(uri, root.FindFolder(uri))

	if err != nil {
		return err
	}

	root.UnknownFiles[uri] = file

	return nil
}

func (root *Root) Trigger(event string) {
	list, exist := root.Listeners[event]

	if !exist {
		return
	}

	for _, cb := range list {
		cb()
	}
}

const RootOnUpdate = "update"

func (root *Root) OnUpdate(cb func()) {
	list, exist := root.Listeners[RootOnUpdate]

	if !exist {
		list = make([]func(), 0)
	}

	root.Listeners[RootOnUpdate] = append(list, cb)
}

func NodeToPosString(node *Node) string {
	return fmt.Sprintf("%d:%d", node.StartByte(), node.EndByte())
}
