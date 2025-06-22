package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
)

func GetDoc(uri Uri) (doc *Doc) {
	doc = root.Docs[uri]

	return
}
