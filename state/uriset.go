package state

import (
	. "github.com/redexp/familymarkup-lsp/types"
)

type UriSet map[Uri]bool

func (uris UriSet) Set(uri Uri) {
	uris[uri] = true
}

func (uris UriSet) Has(uri Uri) bool {
	_, has := uris[uri]

	return has
}
