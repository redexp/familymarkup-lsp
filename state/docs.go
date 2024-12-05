package state

import (
	"iter"
	"os"
	"sync"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	"github.com/redexp/textdocument"
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
)

type Docs map[Uri]*TextDocument

var documents sync.Map

func GetDoc(uri Uri) *TextDocument {
	value, ok := documents.Load(uri)

	if !ok {
		return nil
	}

	return value.(*TextDocument)
}

func (root *Root) OpenDoc(uri Uri) (doc *TextDocument, err error) {
	uri, err = NormalizeUri(uri)

	if err != nil {
		return
	}

	doc = GetDoc(uri)

	if doc != nil {
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

	q, err := familymarkup.GetHighlightQuery()

	if err != nil {
		return
	}

	doc.HighlightCapturesDirty = true

	doc.SetHighlightQuery(q, &textdocument.Ignore{
		Missing: true,
	})

	documents.Store(uri, doc)

	return
}

func GetOpenDocsIter() iter.Seq2[Uri, *TextDocument] {
	return func(yield func(Uri, *TextDocument) bool) {
		documents.Range(func(key, value any) bool {
			return yield(key.(Uri), value.(*TextDocument))
		})
	}
}

func CloseDoc(uri Uri) {
	doc := GetDoc(uri)

	if doc == nil {
		return
	}

	doc.Parser.Close()
	documents.Delete(uri)
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

	doc = GetDoc(uri)

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
