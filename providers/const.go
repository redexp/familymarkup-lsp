package providers

import (
	"github.com/redexp/familymarkup-lsp/state"
	"github.com/redexp/textdocument"
	"github.com/tliron/glsp"
	serv "github.com/tliron/glsp/server"
)

var (
	typesMap textdocument.HighlightLegend
	server   *serv.Server
	root     *state.Root
)

var supportDiagnostics = false

type Ctx = glsp.Context
