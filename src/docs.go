package src

import (
	"github.com/redexp/textdocument"
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
)

var documents map[Uri]*TextDocument = make(map[Uri]*TextDocument)

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

func openDocText(uri Uri, text string, tree *Tree) (*TextDocument, error) {
	doc := textdocument.NewTextDocument(text)
	doc.Tree = tree
	err := doc.SetParser(getParser())

	if err != nil {
		return nil, err
	}

	q, err := familymarkup.GetHighlightQuery()

	if err != nil {
		return nil, err
	}

	doc.SetHighlightQuery(q, &textdocument.Ignore{
		Missing: true,
	})

	documents[uri] = doc

	return doc, nil
}

func closeDoc(uri Uri) {
	delete(documents, uri)
}

func toString(node *Node, doc *TextDocument) string {
	return node.Content([]byte(doc.Text))
}
