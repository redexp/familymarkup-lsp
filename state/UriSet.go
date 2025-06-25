package state

import (
	. "github.com/redexp/familymarkup-lsp/types"
)

type UriSet map[Uri]struct{}

func (uris UriSet) Set(uri Uri) {
	uris[uri] = struct{}{}
}

func (uris UriSet) Has(uri Uri) bool {
	_, ok := uris[uri]
	return ok
}

func (uris UriSet) Empty() bool {
	return len(uris) == 0
}

func (uris UriSet) Remove(uri Uri) {
	delete(uris, uri)
}
