package providers

import (
	"github.com/redexp/familymarkup-lsp/state"
	"github.com/tliron/glsp"
)

var (
	root *state.Root
)

var warnChildrenWithoutRelations = false

type Ctx = glsp.Context
