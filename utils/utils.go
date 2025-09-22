package utils

import (
	urlParser "net/url"
	"path/filepath"
	"slices"
	"strings"

	proto "github.com/tliron/glsp/protocol_3_16"

	. "github.com/redexp/familymarkup-lsp/types"
	fm "github.com/redexp/familymarkup-parser"
)

var FamilyExt = []string{"fml", "family"}
var MarkdownExt = []string{"md", "mdx"}
var AllExt = slices.Concat(FamilyExt, MarkdownExt)

func UriToPath(uri Uri) (string, error) {
	if strings.HasPrefix(uri, "/") {
		return uri, nil
	}

	url, err := urlParser.Parse(uri)

	if err != nil {
		return "", err
	}

	return url.Path, nil
}

func ToUri(path string) Uri {
	if strings.HasPrefix(path, "/") {
		path = "file://" + path
	}

	return path
}

func Ext(path string) string {
	return strings.ToLower(strings.TrimLeft(filepath.Ext(path), "."))
}

func IsFamilyUri(uri Uri) bool {
	return slices.Contains(FamilyExt, Ext(uri))
}

func IsMarkdownUri(uri Uri) bool {
	return slices.Contains(MarkdownExt, Ext(uri))
}

func TokensToStrings(tokens []*fm.Token) []string {
	list := make([]string, len(tokens))

	for i, token := range tokens {
		list[i] = token.Text
	}

	return list
}

func LocToRange(loc fm.Loc) Range {
	return Range{
		Start: Position{
			Line:      uint32(loc.Start.Line),
			Character: uint32(loc.Start.Char),
		},
		End: Position{
			Line:      uint32(loc.End.Line),
			Character: uint32(loc.End.Char),
		},
	}
}

func RangeToLoc(r Range) fm.Loc {
	return fm.Loc{
		Start: fm.Position{
			Line: int(r.Start.Line),
			Char: int(r.Start.Character),
		},
		End: fm.Position{
			Line: int(r.End.Line),
			Char: int(r.End.Character),
		},
	}
}

func PositionToRange(pos Position) Range {
	return Range{
		Start: pos,
		End:   pos,
	}
}

func LocPosToPosition(pos fm.Position) Position {
	return Position{
		Line:      uint32(pos.Line),
		Character: uint32(pos.Char),
	}
}

func TokenToPosition(token *fm.Token) proto.Position {
	return proto.Position{
		Line:      uint32(token.Line),
		Character: uint32(token.Char),
	}
}

func TokenEndToPosition(token *fm.Token) proto.Position {
	return proto.Position{
		Line:      uint32(token.Line),
		Character: uint32(token.EndChar()),
	}
}

func TokenToRange(token *fm.Token) proto.Range {
	return proto.Range{
		Start: TokenToPosition(token),
		End:   TokenEndToPosition(token),
	}
}

func Slice(text string, start, end int) string {
	return string([]rune(text)[start:end])
}

func SliceToEnd(text string, start int) string {
	return string([]rune(text)[start:])
}
