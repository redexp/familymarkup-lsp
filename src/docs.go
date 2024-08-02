package src

import (
	"github.com/redexp/textdocument"
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
)

var documents map[Uri]*TextDocument = make(map[Uri]*TextDocument)

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
