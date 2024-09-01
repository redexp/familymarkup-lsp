package src

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
)

var trees sync.Map
var readingTrees sync.WaitGroup
var lock sync.Mutex

func waitTreesReady() {
	readingTrees.Wait()
}

func setTree(uri Uri, tree *Tree) {
	trees.Store(uri, tree)
}

func getTree(uri Uri) *Tree {
	value, ok := trees.Load(uri)

	if !ok {
		return nil
	}

	return value.(*Tree)
}

func removeTree(uri Uri) {
	trees.Delete(uri)
}

func getTreeText(uri Uri) (tree *Tree, text []byte, err error) {
	uri, err = normalizeUri(uri)

	if err != nil {
		return
	}

	path, err := uriToPath(uri)

	if err != nil {
		return
	}

	text, err = os.ReadFile(path)

	if err != nil {
		return
	}

	tree = getTree(uri)

	if tree != nil {
		return
	}

	tree, err = parseTree(text)

	if err != nil {
		return
	}

	setTree(uri, tree)

	return
}

func parseTree(text []byte) (*sitter.Tree, error) {
	return getParser().Parse(text)
}

func readTreesFromDir(root string, cb func(*Tree, []byte, string) error) error {
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

			tree, text, err := getTreeText(path)

			if err != nil {
				Debugf("getTreeText(%s) error: %s", path, err.Error())
				return
			}

			cb(tree, text, path)
		}()

		return nil
	})
}
