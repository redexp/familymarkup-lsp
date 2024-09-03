package src

import (
	"github.com/redexp/textdocument"
	sitter "github.com/smacker/go-tree-sitter"
	proto "github.com/tliron/glsp/protocol_3_16"
	serv "github.com/tliron/glsp/server"
)

type TextDocument = textdocument.TextDocument
type Uri = proto.DocumentUri
type Tree = sitter.Tree
type Node = sitter.Node
type Position = proto.Position

var (
	typesMap textdocument.HighlightLegend
	server   *serv.Server
	root     *Root
)

var supportDiagnostics = false
