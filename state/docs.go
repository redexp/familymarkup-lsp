package state

import (
	"iter"
	"os"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	"github.com/redexp/textdocument"
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
)

type Docs map[Uri]*TextDocument

var documents Docs = make(Docs)

func (root *Root) OpenDoc(uri Uri) (doc *TextDocument, err error) {
	uri, err = NormalizeUri(uri)

	if err != nil {
		return
	}

	doc, ok := documents[uri]

	if ok {
		return
	}

	tree, text, err := GetTreeText(uri)

	if err != nil {
		return
	}

	return root.OpenDocText(uri, string(text), tree)
}

func (root *Root) OpenDocText(uri Uri, text string, tree *Tree) (doc *TextDocument, err error) {
	doc = textdocument.NewTextDocument(text)
	doc.Tree = tree
	doc.Parser = CreateParser()

	if tree == nil {
		doc.UpdateTree(nil)
		SetTree(uri, doc.Tree)
	}

	SetDocHighlightQuery(doc, root.SurnameFirst)

	documents[uri] = doc

	return
}

func SetDocHighlightQuery(doc *TextDocument, surnameFirst bool) (err error) {
	var q *sitter.Query

	if surnameFirst {
		q, err = familymarkup.GetHighlightQuery()
	} else {
		q, err = familymarkup.GetHighlightQueryLastNameFirst()
	}

	if err != nil {
		return
	}

	doc.HighlightCapturesDirty = true

	doc.SetHighlightQuery(q, &textdocument.Ignore{
		Missing: true,
	})

	return
}

func GetOpenDocsIter() iter.Seq2[Uri, *TextDocument] {
	return func(yield func(Uri, *TextDocument) bool) {
		for uri, doc := range documents {
			if !yield(uri, doc) {
				break
			}
		}
	}
}

func CloseDoc(uri Uri) {
	doc, exist := documents[uri]

	if !exist {
		return
	}

	doc.Parser.Close()
	delete(documents, uri)
}

func RemoveDoc(uri Uri) error {
	uri, err := NormalizeUri(uri)

	if err != nil {
		return err
	}

	CloseDoc(uri)
	RemoveTree(uri)

	return nil
}

func TempDoc(uri Uri) (doc *TextDocument, err error) {
	uri, err = NormalizeUri(uri)

	if err != nil {
		return
	}

	doc = documents[uri]

	if doc != nil {
		return
	}

	tree, text, err := GetTreeText(uri)

	if err != nil {
		return
	}

	doc = textdocument.NewTextDocument(string(text))
	doc.Tree = tree

	return
}

func UriFileExist(uri Uri) bool {
	path, err := UriToPath(uri)

	if err != nil {
		return false
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func ToString(node *Node, doc *TextDocument) string {
	if node == nil {
		return ""
	}

	return node.Content([]byte(doc.Text))
}

func (docs Docs) Get(uri Uri) (doc *TextDocument, err error) {
	doc = docs[uri]

	if doc != nil {
		return
	}

	doc, err = TempDoc(uri)

	if err != nil {
		return
	}

	docs[uri] = doc

	return
}
