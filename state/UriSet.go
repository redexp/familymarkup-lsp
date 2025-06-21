package state

import (
	. "github.com/redexp/familymarkup-lsp/types"
)

type UriSet map[Uri]struct{}

func (uris UriSet) Set(uri Uri) {
	uris[uri] = struct{}{}
}

func (uris UriSet) Remove(uri Uri) {
	delete(uris, uri)
}
