package state

import (
	"os"
	"sync"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	sitter "github.com/smacker/go-tree-sitter"
)

var trees sync.Map

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
