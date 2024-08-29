package src

import (
	"os"

	"github.com/redexp/textdocument"
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
)

type Docs map[Uri]*TextDocument

var documents Docs = make(Docs)

func openDoc(uri Uri) (doc *TextDocument, err error) {
	uri, err = normalizeUri(uri)

	if err != nil {
		return
	}

	doc, ok := documents[uri]

	if ok {
		return
	}

	tree, text, err := getTreeText(uri)

	if err != nil {
		return
	}

	return openDocText(uri, string(text), tree)
}

func openDocText(uri Uri, text string, tree *Tree) (doc *TextDocument, err error) {
	doc = textdocument.NewTextDocument(text)
	doc.Tree = tree
	err = doc.SetParser(getParser())

	if err != nil {
		return
	}

	q, err := familymarkup.GetHighlightQuery()

	if err != nil {
		return
	}

	doc.SetHighlightQuery(q, &textdocument.Ignore{
		Missing: true,
	})

	documents[uri] = doc
	setTree(uri, doc.Tree)

	return
}

func closeDoc(uri Uri) {
	delete(documents, uri)
}

func tempDoc(uri Uri) (doc *TextDocument, err error) {
	uri, err = normalizeUri(uri)

	if err != nil {
		return
	}

	doc = documents[uri]

	if doc != nil {
		return
	}

	tree, text, err := getTreeText(uri)

	if err != nil {
		return
	}

	doc = textdocument.NewTextDocument(string(text))
	doc.Tree = tree

	return
}

func docExist(uri Uri) bool {
	path, err := uriToPath(uri)

	if err != nil {
		return false
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func toString(node *Node, doc *TextDocument) string {
	return node.Content([]byte(doc.Text))
}

func (docs Docs) Get(uri Uri) (doc *TextDocument, err error) {
	doc = docs[uri]

	if doc != nil {
		return
	}

	doc, err = tempDoc(uri)

	if err != nil {
		return
	}

	docs[uri] = doc

	return
}
