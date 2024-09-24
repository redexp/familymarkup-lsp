package state

import (
	"io/fs"
	"math"
	"path/filepath"
	"slices"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
)

func addDuplicate(duplicates Duplicates, name string, dup *Duplicate) {
	_, exist := duplicates[name]

	if !exist {
		duplicates[name] = make([]*Duplicate, 0)
	}

	duplicates[name] = append(duplicates[name], dup)
}

func filterRefs(refs []*Ref, uris UriSet) []*Ref {
	return slices.DeleteFunc(refs, func(ref *Ref) bool {
		return uris.Has(ref.Uri)
	})
}

func getAliasesNode(node *Node) *Node {
	next := node.NextNamedSibling()

	if IsNameAliases(next) {
		return next
	}

	parent := node.Parent()

	if IsNameDef(parent) {
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

func WalkFiles(uri Uri, extensions []string, cb func(Uri, string) error) (err error) {
	rootPath, err := UriToPath(uri)

	if err != nil {
		return
	}

	return filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := Ext(info.Name())

		if !slices.Contains(extensions, ext) {
			return nil
		}

		return cb(ToUri(path), ext)
	})
}
