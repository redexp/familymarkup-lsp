package state

import (
	. "github.com/redexp/familymarkup-lsp/types"
)

const (
	FileCreate = uint8(iota)
	FileChange
	FileRename
	FileDelete
	FileOpen
	FileClose
)

type UriSet map[Uri]uint8

func (uris UriSet) Set(uri Uri) {
	uris[uri] = 0
}

func (uris UriSet) SetState(uri Uri, val uint8) {
	uris[uri] = val
}

func (uris UriSet) Has(uri Uri) bool {
	_, has := uris[uri]

	return has
}

func (uris UriSet) Remove(uri Uri) {
	delete(uris, uri)
}
