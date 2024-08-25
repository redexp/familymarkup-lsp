package src

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
)

type Trees map[Uri]*Tree

var trees Trees = Trees{}

func setTree(uri Uri, tree *Tree) {
	trees[uri] = tree
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

	tree, err = parseTree(text)

	if err != nil {
		return
	}

	setTree(uri, tree)

	return
}

func parseTree(text []byte) (*sitter.Tree, error) {
	return getParser().ParseCtx(context.Background(), nil, text)
}

func readTreesFromDir(root string, cb func(*Tree, []byte, string) error) error {
	var wg sync.WaitGroup

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))

		if ext != ".fm" && ext != ".family" {
			return nil
		}

		wg.Add(1)

		go func() {
			defer wg.Done()

			tree, text, _ := getTreeText(path)

			cb(tree, text, path)
		}()

		return nil
	})

	if err != nil {
		return err
	}

	wg.Wait()

	return nil
}
