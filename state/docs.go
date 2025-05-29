package state

import (
	fm "github.com/redexp/familymarkup-parser"
	"iter"
	"os"
	"slices"
	"sync"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	"github.com/redexp/textdocument"
)

type Doc struct {
	*TextDocument

	Uri    Uri
	Tokens []*fm.Token
	Root   *fm.Root
}

type Docs map[Uri]*Doc

var documents sync.Map

func CreateDoc(uri Uri, text string) *Doc {
	doc := &Doc{
		TextDocument: textdocument.NewTextDocument(text),
		Uri:          uri,
		Tokens:       fm.Lexer(text),
	}

	doc.Root = fm.ParseTokens(doc.Tokens)

	return doc
}

func CreateDocFromUri(uri Uri) (doc *Doc, err error) {
	text, err := GetText(uri)

	if err != nil {
		return
	}

	doc = CreateDoc(uri, text)

	return
}

func GetDoc(uri Uri) *Doc {
	value, ok := documents.Load(uri)

	if !ok {
		return nil
	}

	return value.(*Doc)
}

func (root *Root) OpenDoc(uri Uri) (doc *Doc, err error) {
	uri, err = NormalizeUri(uri)

	if err != nil {
		return
	}

	doc = GetDoc(uri)

	if doc != nil {
		return
	}

	text, err := GetText(uri)

	if err != nil {
		return
	}

	return root.OpenDocText(uri, text)
}

func (root *Root) OpenDocText(uri Uri, text string) (doc *Doc, err error) {
	doc = CreateDoc(uri, text)

	documents.Store(uri, doc)

	return
}

func GetOpenDocsIter() iter.Seq2[Uri, *Doc] {
	return func(yield func(Uri, *Doc) bool) {
		documents.Range(func(key, value any) bool {
			return yield(key.(Uri), value.(*Doc))
		})
	}
}

func CloseDoc(uri Uri) {
	doc := GetDoc(uri)

	if doc == nil {
		return
	}

	documents.Delete(uri)
}

func RemoveDoc(uri Uri) error {
	uri, err := NormalizeUri(uri)

	if err != nil {
		return err
	}

	CloseDoc(uri)

	return nil
}

func TempDoc(uri Uri) (doc *Doc, err error) {
	uri, err = NormalizeUri(uri)

	if err != nil {
		return
	}

	doc = GetDoc(uri)

	if doc != nil {
		return
	}

	return CreateDocFromUri(uri)
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

func GetText(uri Uri) (text string, err error) {
	uri, err = NormalizeUri(uri)

	if err != nil {
		return
	}

	path, err := UriToPath(uri)

	if err != nil {
		return
	}

	bytes, err := os.ReadFile(path)

	if err != nil {
		return
	}

	text = string(bytes)

	return
}

func (docs Docs) Get(uri Uri) (doc *Doc, err error) {
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

func (doc *Doc) GetTextByLoc(loc fm.Loc) (string, error) {
	r := LocToRange(loc)
	return doc.GetTextByRange(&r)
}

func (doc *Doc) TokenIndex(token *fm.Token) int {
	return slices.Index(doc.Tokens, token)
}

func (doc *Doc) PrevNextTokens(token *fm.Token) (prev *fm.Token, next *fm.Token) {
	count := len(doc.Tokens)

	if count <= 1 {
		return
	}

	index := doc.TokenIndex(token)

	if index > 0 {
		prev = doc.Tokens[index-1]
	}

	if index < count-1 {
		next = doc.Tokens[index+1]
	}

	return
}

func (doc *Doc) PrevNextNonSpaceTokens(token *fm.Token) (prev *fm.Token, next *fm.Token) {
	count := len(doc.Tokens)

	if count <= 1 {
		return
	}

	index := doc.TokenIndex(token)

	for i := index - 1; i >= 0; i-- {
		t := doc.Tokens[i]

		if t.Type != fm.TokenSpace {
			prev = t
			break
		}
	}

	for i := index + 1; i < count; i++ {
		t := doc.Tokens[i]

		if t.Type != fm.TokenSpace {
			next = t
			break
		}
	}

	return
}

func (doc *Doc) GetTokenByPosition(pos *Position) *fm.Token {
	offset, err := doc.PositionToByteIndex(pos)

	if err != nil {
		return nil
	}

	for _, token := range doc.Tokens {
		if uint32(token.Offest) < offset {
			continue
		}

		return token
	}

	return nil
}

func (doc *Doc) FindFamilyByLoc(loc fm.Loc) *fm.Family {
	for _, f := range doc.Root.Families {
		switch f.OverlapType(loc) {
		case fm.OverlapAfter:
			return nil
		case fm.OverlapOuter:
			return f
		}
	}

	return nil
}

func (doc *Doc) FindFamilyByRange(r Range) *fm.Family {
	return doc.FindFamilyByLoc(RangeToLoc(r))
}

func (doc *Doc) FindRelationByRange(r Range) *fm.Relation {
	loc := RangeToLoc(r)

	for _, f := range doc.Root.Families {
		for _, rel := range f.Relations {
			switch rel.OverlapType(loc) {
			case fm.OverlapAfter:
				return nil
			case fm.OverlapOuter:
				return rel
			}
		}
	}

	return nil
}

func (doc *Doc) FindPersonByRange(r Range) *fm.Person {
	loc := RangeToLoc(r)

	for _, f := range doc.Root.Families {
		for _, rel := range f.Relations {
			for _, p := range rel.Sources.Persons {
				switch p.OverlapType(loc) {
				case fm.OverlapBefore:
					continue
				case fm.OverlapAfter:
					return nil
				default:
					return p
				}
			}

			if rel.Targets == nil {
				continue
			}

			for _, p := range rel.Sources.Persons {
				switch p.OverlapType(loc) {
				case fm.OverlapBefore:
					continue
				case fm.OverlapAfter:
					return nil
				default:
					return p
				}
			}
		}
	}

	return nil
}
