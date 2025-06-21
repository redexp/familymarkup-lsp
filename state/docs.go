package state

import (
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"
	"os"
	"slices"
	"strings"
)

type Doc struct {
	Uri    Uri
	Text   string
	Open   bool
	Tokens []*fm.Token
	Root   *fm.Root

	TokensByLine map[int][]*fm.Token
}

type Docs map[Uri]*Doc

func CreateDoc(uri Uri, text string) *Doc {
	doc := &Doc{
		Uri: uri,
	}

	doc.SetText(text)

	return doc
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

func (doc *Doc) SetText(text string) {
	doc.Text = text
	doc.Tokens = fm.Lexer(text)
	doc.Root = fm.ParseTokens(doc.Tokens)

	doc.TokensByLine = make(map[int][]*fm.Token)

	for _, token := range doc.Tokens {
		doc.TokensByLine[token.Line] = append(doc.TokensByLine[token.Line], token)
	}
}

func (doc *Doc) GetTextByLine(line int) string {
	tokens, ok := doc.TokensByLine[line]

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
		tokens, ok := doc.TokensByLine[i]

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

func (doc *Doc) GetTrimTokensByLine(line int) (tokens []*fm.Token) {
	tokens, ok := doc.TokensByLine[line]

	if !ok {
		return []*fm.Token{}
	}

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

		if token.Type == fm.TokenSpace || token.SubType == fm.TokenNL {
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

func (doc *Doc) GetTokenByPosition(pos Position) *fm.Token {
	line := int(pos.Line)
	char := int(pos.Character)

	tokens, ok := doc.TokensByLine[line]

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

func (doc *Doc) PosToOffset(pos Position) int {
	line := int(pos.Line)
	char := int(pos.Character)

	list, ok := doc.TokensByLine[line]

	if !ok {
		return len(doc.Text)
	}

	for _, token := range list {
		if token.EndChar() < char {
			continue
		}

		return token.Offest + len(Slice(token.Text, 0, char-token.Char))
	}

	count := len(list)
	last := list[count-1]

	return last.End()
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
