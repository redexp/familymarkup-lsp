package state

import (
	fm "github.com/redexp/familymarkup-parser"
	"iter"
	"os"
	"slices"
	"strings"
	"sync"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
)

type Doc struct {
	Uri    Uri
	Text   string
	Tokens []*fm.Token
	Root   *fm.Root

	TokensByLines map[int][]*fm.Token
}

type Docs map[Uri]*Doc

var documents sync.Map

func CreateDoc(uri Uri, text string) *Doc {
	doc := &Doc{
		Uri: uri,
	}

	doc.SetText(text)

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

func (doc *Doc) SetText(text string) {
	doc.Text = text
	doc.Tokens = fm.Lexer(text)
	doc.Root = fm.ParseTokens(doc.Tokens)

	doc.TokensByLines = make(map[int][]*fm.Token)

	for _, token := range doc.Tokens {
		doc.TokensByLines[token.Line] = append(doc.TokensByLines[token.Line], token)
	}
}

func (doc *Doc) GetTextByLine(line int) string {
	tokens, ok := doc.TokensByLines[line]

	if !ok {
		return ""
	}

	var b strings.Builder

	for _, token := range tokens {
		b.WriteString(token.Text)
	}

	return b.String()
}

func (doc *Doc) GetTextByRange(r Range) string {
	return doc.GetTextByLoc(RangeToLoc(r))
}

func (doc *Doc) GetTextByLoc(loc fm.Loc) string {
	var s strings.Builder

	for i := loc.Start.Line; i <= loc.End.Line; i++ {
		tokens, ok := doc.TokensByLines[i]

		if !ok {
			continue
		}

		for _, token := range tokens {
			switch loc.OverlapType(token.Loc()) {
			case fm.OverlapBefore:
				return s.String()
			case fm.OverlapAfter:
				continue
			case fm.OverlapByStart:
				s.WriteString(SliceToEnd(token.Text, loc.Start.Char-token.Char))
			case fm.OverlapOuter:
				s.WriteString(token.Text)
			case fm.OverlapInner:
				s.WriteString(Slice(token.Text, loc.Start.Char-token.Char, loc.End.Char-token.Char))
			case fm.OverlapByEnd:
				s.WriteString(Slice(token.Text, 0, loc.End.Char-token.Char))
			}
		}
	}

	return s.String()
}

func (doc *Doc) TokenIndex(token *fm.Token) int {
	return slices.Index(doc.Tokens, token)
}

func (doc *Doc) GetTokensByLine(line int) (tokens []*fm.Token) {
	start := -1
	end := -1

	for i, token := range doc.Tokens {
		if token.Line < line {
			continue
		}

		if token.Line > line {
			end = i
			break
		}

		if start == -1 {
			start = i
		}
	}

	if start == -1 {
		return
	}

	if end == -1 {
		return doc.Tokens[start:]
	}

	return doc.Tokens[start:end]
}

func (doc *Doc) GetTrimTokensByLine(line int) (tokens []*fm.Token) {
	tokens = doc.GetTokensByLine(line)
	count := len(tokens)

	if count == 0 {
		return
	}

	start := 0
	end := count

	for _, token := range tokens {
		if token.Type == fm.TokenSpace {
			start++
		} else {
			break
		}
	}

	for i := count - 1; i >= 0; i-- {
		token := tokens[i]

		if token.Type == fm.TokenSpace {
			end--
		} else {
			break
		}
	}

	return tokens[start:end]
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
	line := int(pos.Line)
	char := int(pos.Character)

	tokens, ok := doc.TokensByLines[line]

	if !ok {
		return nil
	}

	for _, token := range tokens {
		if token.IsOnPosition(line, char) {
			return token
		}
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

func (doc *Doc) FindPersonByLine(line int) *fm.Person {
	return doc.FindPersonByRange(Range{
		Start: Position{
			Line: uint32(line),
		},
		End: Position{
			Line: uint32(line),
		},
	})
}
