package providers

import (
	"github.com/redexp/familymarkup-lsp/state"
	"github.com/tliron/glsp"
	serv "github.com/tliron/glsp/server"
)

var (
	server *serv.Server
	root   *state.Root
)

var supportDiagnostics = false
var warnChildrenWithoutRelations = false

type Ctx = glsp.Context
