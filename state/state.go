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
	DirtyUris    DirtyUris
	Labels       map[Uri][]string
	Listeners    Listeners

	UpdateLock sync.Mutex
}

func CreateRoot() *Root {
	return &Root{
		Folders:      make(UriSet),
		Docs:         make(Docs),
		Families:     make(Families),
		Duplicates:   make(Duplicates),
		NodeRefs:     make(NodeRefs),
		UnknownRefs:  make([]*Ref, 0),
		UnknownFiles: make(Files),
		DirtyUris:    make(DirtyUris),
		Labels:       make(map[Uri][]string),
		Listeners:    make(Listeners),
	}
}

func (root *Root) SetFolders(folders []Uri) {
	root.Folders = make(UriSet)

	for _, uri := range folders {
		root.Folders.Set(uri)
	}

	type TextTree struct {
		Uri  Uri
		Text string
	}

	textTrees := make(chan TextTree, 3)

	go func() {
		for uri := range root.Folders {
			_ = WalkFiles(uri, AllExt, func(uri Uri, ext string) error {
				if slices.Contains(MarkdownExt, ext) {
					textTrees <- TextTree{
						Uri: uri,
					}

					return nil
				}

				text, err := GetText(uri)

				if err != nil {
					return err
				}

				textTrees <- TextTree{
					Uri:  uri,
					Text: text,
				}

				return nil
			})
		}

		close(textTrees)
	}()

	for item := range textTrees {
		root.DirtyUris.SetText(item.Uri, UriCreate, item.Text)
	}
}

func (root *Root) OpenDoc(uri Uri) (doc *Doc) {
	doc, ok := root.Docs[uri]

	if !ok {
		return
	}

	doc.Open = true

	return
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
						Type:  RefTypeSurname,
						Uri:   uri,
						Token: person.Surname,
					})
				}

				if person.IsChild {
					mem := family.AddMember(person)

					root.AddRef(&Ref{
						Type:   RefTypeName,
						Uri:    uri,
						Member: mem,
						Person: person,
					})

					continue
				}

				if person.Side == fm.SideSources && rel.IsFamilyDef {
					if person.Surname == nil {
						mem := family.FindMember(person.Name.Text)

						if mem == nil {
							mem = family.AddMember(person)
						}

						root.AddRef(&Ref{
							Type:   RefTypeName,
							Uri:    uri,
							Member: mem,
							Person: person,
						})
					} else {
						mem := family.AddMember(person)

						root.AddRef(&Ref{
							Type:   RefTypeOrigin,
							Uri:    uri,
							Member: mem,
							Person: person,
						})
					}

					continue
				}

				if person.Surname == nil {
					root.AddRef(&Ref{
						Type:   RefTypeName,
						Uri:    uri,
						Family: family,
						Person: person,
					})
				} else {
					root.AddRef(&Ref{
						Type:   RefTypeNameSurname,
						Uri:    uri,
						Person: person,
					})
				}
			}
		}
	}
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
	root.DirtyUris = make(DirtyUris)

	root.UnknownRefs = slices.DeleteFunc(root.UnknownRefs, func(ref *Ref) bool {
		return uris.Has(ref.Uri)
	})

	// update markdown files
	for uri, item := range uris {
		delete(root.Docs, uri)
		delete(root.Labels, uri)
		delete(root.NodeRefs, uri)

		if !IsMarkdownUri(uri) {
			continue
		}

		uris.Remove(uri)

		if _, ok := root.UnknownFiles[uri]; ok {
			if item.IsDeleted() {
				delete(root.UnknownFiles, uri)
			}

			continue
		}

		if item.IsDeleted() {
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

	for family := range root.FamilyIter() {
		if !uris.Has(family.Uri) {
			continue
		}

		for member := range family.MembersIter() {
			if member.InfoUri != "" {
				err = root.AddUnknownFile(member.InfoUri)

				if err != nil {
					return
				}
			}
		}

		root.RemoveFamily(family)
	}

	for uri, item := range uris {
		if !IsFamilyUri(uri) || item.IsDeleted() {
			continue
		}

		doc := CreateDoc(uri, item.Text)
		doc.Open = item.State == UriOpen

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

	root.AddNodeRef(family.Uri, &Ref{
		Type:   RefTypeSurname,
		Family: family,
		Token:  family.Node.Name,
	})

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

	set := make(map[*Member]struct{})

	for member := range f.MembersIter() {
		set[member] = struct{}{}
	}

	for uri, refs := range root.NodeRefs {
		for pos, ref := range refs {
			switch ref.Type {
			case RefTypeName, RefTypeNameSurname:
				if _, ok := set[ref.Member]; ok {
					delete(refs, pos)
					ref.Member = nil
					root.AddUnknownRef(ref)
				}

			case RefTypeSurname:
				if ref.Family == f {
					delete(refs, pos)
					ref.Family = nil
					root.AddUnknownRef(ref)
				}

			case RefTypeOrigin:
				if _, ok := set[ref.Member.Origin]; ok {
					delete(refs, pos)
					ref.Member.Origin = nil
					root.AddUnknownRef(ref)
				}
			}
		}

		if len(refs) == 0 {
			delete(root.NodeRefs, uri)
		}
	}
}

func (root *Root) FamilyIter() iter.Seq[*Family] {
	return func(yield func(*Family) bool) {
		check := createUniqYield(yield)

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

func (root *Root) FamiliesByUriIter(uri Uri) iter.Seq[*Family] {
	return func(yield func(*Family) bool) {
		for family := range root.FamilyIter() {
			if family.Uri != uri {
				continue
			}

			if !yield(family) {
				return
			}
		}
	}
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
		ref.Family = root.FindFamily(ref.Token.Text)

		if ref.Family != nil {
			root.AddNodeRef(ref.Uri, ref)
		} else {
			root.AddUnknownRef(ref)
		}

	case RefTypeName:
		ref.Token = ref.Person.Name

		mem := ref.Member

		if mem == nil {
			mem = ref.Family.FindMember(ref.Token.Text)
		}

		if mem == nil {
			root.AddUnknownRef(ref)
			return
		}

		ref.Member = mem

		root.AddNodeRef(ref.Uri, ref)

	case RefTypeNameSurname:
		p := ref.Person
		_, mem := root.FindMember(p.Surname.Text, p.Name.Text)

		if mem == nil {
			root.AddUnknownRef(ref)
			return
		}

		ref.Member = mem
		ref.Token = p.Name

		root.AddNodeRef(ref.Uri, ref)

	case RefTypeOrigin:
		mem := ref.Member

		f, origin := root.FindMember(mem.Surname, mem.Name)

		if origin == nil {
			root.AddUnknownRef(ref)
			return
		}

		dups, ok := f.Duplicates[mem.Name]

		if ok {
			for _, dup := range dups {
				surname := dup.Member.Surname

				if surname == "" {
					continue
				}

				f = root.FindFamily(surname)

				if f == mem.Family {
					origin = dup.Member
					break
				}
			}
		}

		mem.Origin = origin
		ref.Token = mem.Person.Name
		root.AddNodeRef(ref.Uri, ref)
	}
}

func (root *Root) AddNodeRef(uri Uri, ref *Ref) {
	_, exist := root.NodeRefs[uri]

	if !exist {
		root.NodeRefs[uri] = make(map[string]*Ref)
	}

	pos := TokenToPosString(ref.Token)

	root.NodeRefs[uri][pos] = ref
}

func (root *Root) GetRefByToken(uri Uri, token *fm.Token) *Ref {
	nodesMap, exist := root.NodeRefs[uri]

	if !exist {
		return nil
	}

	ref, exist := nodesMap[TokenToPosString(token)]

	if !exist {
		return nil
	}

	return ref
}

func (root *Root) GetRefByPosition(uri Uri, pos Position) *Ref {
	nodesMap, exist := root.NodeRefs[uri]

	if !exist {
		return nil
	}

	line := int(pos.Line)
	char := int(pos.Character)

	var endRef *Ref

	for _, ref := range nodesMap {
		token := ref.Token

		if token.Line != line {
			continue
		}

		end := token.EndChar()

		if token.Char <= char && char < end {
			return ref
		}

		if char == end {
			endRef = ref
		}
	}

	return endRef
}

func (root *Root) GetMemberByToken(uri Uri, token *fm.Token) *Member {
	ref := root.GetRefByToken(uri, token)

	if ref == nil {
		return nil
	}

	return ref.Member
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
			if u.Token == ref.Token {
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

func (root *Root) RefsIter() iter.Seq2[*Ref, Uri] {
	return func(yield func(*Ref, Uri) bool) {
		for uri, refs := range root.NodeRefs {
			for _, ref := range refs {
				if !yield(ref, uri) {
					return
				}
			}
		}
	}
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
