package state

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	sitter "github.com/smacker/go-tree-sitter"
)

var trees sync.Map
var readingTrees sync.WaitGroup

func WaitTreesReady() {
	readingTrees.Wait()
}

func SetTree(uri Uri, tree *Tree) {
	trees.Store(uri, tree)
}

func GetTree(uri Uri) *Tree {
	value, ok := trees.Load(uri)

	if !ok {
		return nil
	}

	return value.(*Tree)
}

func RemoveTree(uri Uri) {
	trees.Delete(uri)
}

func GetTreeText(uri Uri) (tree *Tree, text []byte, err error) {
	uri, err = NormalizeUri(uri)

	if err != nil {
		return
	}

	path, err := UriToPath(uri)

	if err != nil {
		return
	}

	text, err = os.ReadFile(path)

	if err != nil {
		return
	}

	tree = GetTree(uri)

	if tree != nil {
		return
	}

	tree, err = ParseTree(text)

	if err != nil {
		return
	}

	SetTree(uri, tree)

	return
}

func ParseTree(text []byte) (*sitter.Tree, error) {
	return GetParser().Parse(text)
}

func ReadTreesFromDir(root string, cb func(*Tree, []byte, string) error) error {
	return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))

		if ext != ".fm" && ext != ".family" {
			return nil
		}

		readingTrees.Add(1)

		go func() {
			defer readingTrees.Done()

			tree, text, err := GetTreeText(path)

			if err != nil {
				return
			}

			cb(tree, text, path)
		}()

		return nil
	})
}
