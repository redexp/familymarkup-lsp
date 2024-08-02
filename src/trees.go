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

func getTreeText(uri Uri) (*Tree, []byte, error) {
	path, err := uriToPath(uri)

	if err != nil {
		return nil, nil, err
	}

	text, err := os.ReadFile(path)

	if err != nil {
		return nil, nil, err
	}

	tree, err := parseTree(text)

	if err != nil {
		return nil, nil, err
	}

	setTree(uri, tree)

	return tree, text, nil
}

func parseTree(text []byte) (*sitter.Tree, error) {
	return getParser().ParseCtx(context.Background(), nil, text)
}

func readTreesFromDir(dir Uri, cb func(*Tree, []byte) error) error {
	root, err := uriToPath(dir)

	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	err = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
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

			cb(tree, text)
		}()

		return nil
	})

	if err != nil {
		return err
	}

	wg.Wait()

	return nil
}
