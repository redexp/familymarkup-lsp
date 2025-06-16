package state

import (
	"fmt"
	fm "github.com/redexp/familymarkup-parser"
	"iter"
	"slices"
	"strings"
	"sync"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
)

type Root struct {
	Folders      UriSet
	Docs         Docs
	Families     Families
	Duplicates   Duplicates
	NodeRefs     NodeRefs
	UnknownRefs  []*Ref
	UnknownFiles Files
	DirtyUris    UriSet
	Labels       map[Uri][]string
	Listeners    Listeners
	Log          func(string, ...any)

	UpdateLock sync.Mutex
}

func CreateRoot(logger func(string, ...any)) *Root {
	return &Root{
		Folders:      make(UriSet),
		Docs:         make(Docs),
		Families:     make(Families),
		Duplicates:   make(Duplicates),
		NodeRefs:     make(NodeRefs),
		UnknownRefs:  make([]*Ref, 0),
		UnknownFiles: make(Files),
		DirtyUris:    make(UriSet),
		Labels:       make(map[Uri][]string),
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
		Doc *Doc
		Uri Uri
		MD  bool
	}

	textTrees := make(chan TextTree, 3)

	go func() {
		for uri := range root.Folders {
			_ = WalkFiles(uri, AllExt, func(uri Uri, ext string) error {
				if slices.Contains(MarkdownExt, ext) {
					textTrees <- TextTree{
						Uri: uri,
						MD:  true,
					}
					return nil
				}

				doc, err := CreateDocFromUri(uri)

				if err != nil {
					return err
				}

				textTrees <- TextTree{
					Doc: doc,
					Uri: uri,
				}

				return nil
			})
		}

		close(textTrees)
	}()

	for item := range textTrees {
		if item.Doc != nil {
			root.Docs[item.Uri] = item.Doc
		}

		root.DirtyUris.SetState(item.Uri, FileCreate)
	}

	return root.UpdateDirty()
}

func (root *Root) EnsureDoc(uri Uri) (doc *Doc, err error) {
	doc, ok := root.Docs[uri]

	if ok {
		return
	}

	text, err := GetText(uri)

	if err != nil {
		return
	}

	doc = CreateDoc(uri, text)

	root.Docs[uri] = doc
	root.DirtyUris.Set(uri)

	return
}

func (root *Root) OpenDoc(uri Uri) (doc *Doc, err error) {
	doc, err = root.EnsureDoc(uri)

	if err != nil {
		return
	}

	doc.Open = true

	return
}

func (root *Root) OpenDocText(uri Uri, text string) *Doc {
	doc, ok := root.Docs[uri]

	if !ok {
		doc = CreateDoc(uri, text)
		doc.Open = true
		root.Docs[uri] = doc
		root.DirtyUris.Set(uri)

		return doc
	}

	doc.Open = true

	if doc.Text != text {
		doc.SetText(text)
		root.DirtyUris.Set(uri)
	}

	return doc
}

func (root *Root) CloseDoc(uri Uri) {
	doc, ok := root.Docs[uri]

	if ok {
		doc.Open = false
	}
}

func (root *Root) Update(doc *Doc) {
	root.Docs[doc.Uri] = doc

	uri := doc.Uri

	var family *Family

	for _, f := range doc.Root.Families {
		family = root.AddFamily(uri, f)

		for _, rel := range f.Relations {
			if rel.IsFamilyDef && rel.Label != nil {
				root.AddLabel(uri, rel.Label.Text)
			}

			for person := range rel.PersonsIter() {
				if person.Unknown != nil || person.Name == nil {
					continue
				}

				if person.Surname != nil {
					root.AddRef(&Ref{
						Type:    RefTypeSurname,
						Uri:     uri,
						Surname: person.Surname,
					})

					if !person.IsChild {
						root.AddRef(&Ref{
							Type:   RefTypeNameSurname,
							Uri:    uri,
							Person: person,
							Family: family,
						})
					}
				}

				if !rel.IsFamilyDef || (person.Side == fm.SideSources && person.Surname == nil && family.HasMember(person.Name.Text)) {
					root.AddRef(&Ref{
						Type:   RefTypeName,
						Uri:    uri,
						Person: person,
						Family: family,
					})
					continue
				}

				mem := family.AddMember(person)

				if person.IsChild && person.Surname != nil {
					// create a member in this surname
					root.AddRef(&Ref{
						Type:   RefTypeOrigin,
						Uri:    uri,
						Origin: mem,
					})
				}
			}
		}
	}

	root.UpdateUnknownRefs()
}

func (root *Root) UpdateUnknownRefs() {
	if len(root.UnknownRefs) == 0 {
		return
	}

	list := root.UnknownRefs
	root.UnknownRefs = Refs{}

	for _, ref := range list {
		root.AddRef(ref)
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

func (root *Root) UpdateDirty() (err error) {
	root.UpdateLock.Lock()
	defer root.UpdateLock.Unlock()

	if len(root.DirtyUris) == 0 {
		return
	}

	uris := root.DirtyUris
	root.DirtyUris = UriSet{}
	root.UnknownRefs = filterRefs(root.UnknownRefs, uris)

	// update markdown files
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
			err = root.AddUnknownFile(uri)

			if err != nil {
				return
			}
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
					err = root.AddUnknownFile(member.InfoUri)

					if err != nil {
						return
					}
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

	for uri, state := range uris {
		delete(root.Labels, uri)

		if state == FileDelete {
			delete(root.Docs, uri)
			continue
		}

		doc, err := root.EnsureDoc(uri)

		if err != nil {
			return err
		}

		root.Update(doc)
	}

	root.UpdateUnknownRefs()
	root.UpdateUnknownFiles()
	root.Trigger(RootOnUpdate)

	return
}

func (root *Root) AddFamily(uri Uri, node *fm.Family) *Family {
	family := &Family{
		Name:       node.Name.Text,
		Aliases:    TokensToStrings(node.Aliases),
		Members:    make(Members),
		Duplicates: make(Duplicates),
		Uri:        uri,
		Node:       node,
		Root:       root,
	}

	names := append(family.Aliases, family.Name)

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

	root.AddNodeRef(family.Uri, &FamMem{Family: family, Token: family.Node.Name})

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
					Type:    RefTypeSurname,
					Uri:     uri,
					Surname: item.Token,
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
	switch ref.Type {
	case RefTypeSurname:
		f := root.FindFamily(ref.Surname.Text)

		if f != nil {
			root.AddNodeRef(ref.Uri, &FamMem{Family: f, Token: ref.Surname})
		} else {
			root.AddUnknownRef(ref)
		}

	case RefTypeName, RefTypeNameSurname:
		name := ref.Person.Name.Text

		f := ref.Family

		if ref.Type == RefTypeNameSurname {
			f = root.FindFamily(ref.Person.Surname.Text)

			if f == nil {
				root.AddUnknownRef(ref)
				return
			}
		}

		mem := f.FindMember(name)

		if mem == nil {
			root.AddUnknownRef(ref)
			return
		}

		dups, exist := f.Duplicates[mem.NormalizeName(name)]

		if exist && ref.Family != nil {
			for _, dup := range dups {
				if dup.Member.Surname == "" {
					continue
				}

				fam := root.FindFamily(dup.Member.Surname)

				if fam == ref.Family {
					mem = dup.Member
					break
				}
			}
		}

		mem.Refs = append(mem.Refs, ref)

		root.AddNodeRef(ref.Uri, &FamMem{Member: mem, Person: ref.Person, Token: ref.Person.Name})

		if ref.Type == RefTypeNameSurname {
			root.AddNodeRef(ref.Uri, &FamMem{Family: f, Person: ref.Person, Token: ref.Person.Surname})
		}

	case RefTypeOrigin:
		origin := ref.Origin

		f := root.FindFamily(origin.Surname)

		if f == nil {
			root.AddUnknownRef(&Ref{
				Type:    RefTypeSurname,
				Uri:     ref.Uri,
				Surname: origin.Person.Surname,
			})

			root.AddUnknownRef(ref)
			return
		}

		var person *fm.Person

		// find the first oRef of this member in that file (usually as mather in family relation)
		for _, oRef := range origin.Refs {
			if oRef.Uri == f.Uri {
				person = oRef.Person
				break
			}
		}

		if person == nil {
			root.AddUnknownRef(ref)
			return
		}

		mem := f.AddMemberName(person, origin.Name, origin.Aliases, "")
		mem.Origin = origin
	}
}

func (root *Root) AddNodeRef(uri Uri, famMem *FamMem) {
	_, exist := root.NodeRefs[uri]

	if !exist {
		root.NodeRefs[uri] = make(map[string]*FamMem)
	}

	pos := TokenToPosString(famMem.Token) // TODO: replace with [line][char]

	_, exist = root.NodeRefs[uri][pos]

	if exist {
		return
	}

	if famMem.Member != nil && famMem.Member.Origin != nil {
		famMem = &FamMem{Member: famMem.Member.Origin}
	}

	root.NodeRefs[uri][pos] = famMem
}

func (root *Root) GetFamMem(uri Uri, token *fm.Token) *FamMem {
	nodesMap, exist := root.NodeRefs[uri]

	if !exist {
		return nil
	}

	famMem, exist := nodesMap[TokenToPosString(token)]

	if !exist {
		return nil
	}

	return famMem
}

func (root *Root) GetFamMemByPosition(uri Uri, pos Position) *FamMem {
	nodesMap, exist := root.NodeRefs[uri]

	if !exist {
		return nil
	}

	line := int(pos.Line)
	char := int(pos.Character)

	for _, famMem := range nodesMap {
		if famMem.Token.IsOnPosition(line, char) {
			return famMem
		}
	}

	return nil
}

func (root *Root) GetMemberByUriToken(uri Uri, token *fm.Token) *Member {
	famMem := root.GetFamMem(uri, token)

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
	if ref.Type == RefTypeSurname {
		for _, u := range root.UnknownRefs {
			if u.Surname == ref.Surname {
				return
			}
		}
	}

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

func (root *Root) AddLabel(uri Uri, label string) {
	label = strings.TrimSpace(label)

	list, exist := root.Labels[uri]

	if !exist {
		root.Labels[uri] = []string{label}
		return
	}

	if slices.Contains(list, label) {
		return
	}

	root.Labels[uri] = append(list, label)
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

func TokenToPosString(token *fm.Token) string {
	return fmt.Sprintf("%d:%d", token.Line, token.Char)
}
