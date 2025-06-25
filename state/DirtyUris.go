package state

import (
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	"strings"
)

type DirtyUris map[Uri]*TextState

type TextState struct {
	State UriState
	Text  string
}

type UriState uint8

const (
	UriCreate UriState = 1 + iota
	UriOpen
	UriChange
	UriDelete
)

func (uris DirtyUris) Set(uri Uri, state UriState) {
	uris[uri] = &TextState{
		State: state,
	}
}

func (uris DirtyUris) SetText(uri Uri, state UriState, text string) {
	uris[uri] = &TextState{
		State: state,
		Text:  text,
	}
}

func (uris DirtyUris) ChangeText(doc *Doc, r *Range, newText string) {
	uri := doc.Uri

	if !uris.Has(uri) {
		uris[uri] = &TextState{
			State: UriChange,
			Text:  doc.Text,
		}
	}

	loc := RangeToLoc(*r)
	text := uris[uri].Text

	offsetStart := getLineOffset(text, loc.Start.Line, 0)
	offsetEnd := getLineOffset(text, loc.End.Line-loc.Start.Line, offsetStart)

	prefix := text[:offsetStart] + Slice(text[offsetStart:], 0, loc.Start.Char)
	suffix := SliceToEnd(text[offsetEnd:], loc.End.Char)

	uris[uri].Text = prefix + newText + suffix
}

func (uris DirtyUris) Has(uri Uri) bool {
	_, has := uris[uri]

	return has
}

func (uris DirtyUris) Remove(uri Uri) {
	delete(uris, uri)
}

func (uris DirtyUris) GetDeleted() UriSet {
	list := UriSet{}

	for uri, item := range uris {
		if item.IsDeleted() {
			list.Set(uri)
		}
	}

	return list
}

func (item *TextState) IsDeleted() bool {
	return item.State == UriDelete
}

func getLineOffset(text string, line int, offset int) int {
	if line == 0 {
		return offset
	}

	c := "\n"
	n := 0

	for {
		i := strings.Index(text[offset:], c)

		if i == -1 {
			return len(text)
		}

		offset += i + 1

		n++

		if n == line {
			return offset
		}
	}
}
