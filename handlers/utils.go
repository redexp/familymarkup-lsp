package handlers

import (
	"context"
	"encoding/json"
	"net/url"
	"os"

	"github.com/redexp/textdocument"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tliron/glsp"
	serv "github.com/tliron/glsp/server"
)

func CreateServer(handlers glsp.Handler) {
	server = serv.NewServer(handlers, "familymarkup", false)
	server.RunStdio()
}

func logDebug(msg string, data any) {
	if server == nil || server.Log.GetMaxLevel() < 2 {
		return
	}

	str, _ := json.MarshalIndent(data, "", "  ")
	server.Log.Debugf(msg, str)
}

func getTree(uri Uri) (*sitter.Tree, error) {
	doc, ok := documents[uri]

	if ok && doc.Tree != nil {
		return doc.Tree, nil
	}

	tree, _, err := getTreeText(uri)

	return tree, err
}

func getTreeText(uri Uri) (*Tree, []byte, error) {
	tree, ok := trees[uri]

	if ok {
		return tree, nil, nil
	}

	u, err := url.Parse(uri)

	if err != nil {
		return nil, nil, err
	}

	src, err := os.ReadFile(u.Path)

	if err != nil {
		return nil, nil, err
	}

	tree, err = parser.ParseCtx(context.Background(), nil, src)

	if err != nil {
		return nil, nil, err
	}

	trees[uri] = tree

	return tree, src, nil
}

func openDoc(uri Uri) (*TextDocument, error) {
	doc, ok := documents[uri]

	if ok {
		return doc, nil
	}

	tree, text, err := getTreeText(uri)

	if err != nil {
		return nil, err
	}

	return openDocText(uri, string(text), tree)
}

func openDocText(uri Uri, text string, tree *Tree) (*TextDocument, error) {
	doc := textdocument.NewTextDocument(text)
	doc.Tree = tree
	err := doc.SetParser(parser)

	if err != nil {
		return nil, err
	}

	documents[uri] = doc

	return doc, nil
}

func someError(list ...error) bool {
	for _, err := range list {
		if err != nil {
			return true
		}
	}

	return false
}

func findError(list ...error) error {
	for _, err := range list {
		if err != nil {
			return err
		}
	}

	return nil
}
