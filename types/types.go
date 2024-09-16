package types

import (
	"github.com/redexp/textdocument"
	sitter "github.com/smacker/go-tree-sitter"
	proto "github.com/tliron/glsp/protocol_3_16"
)

type TextDocument = textdocument.TextDocument
type Uri = proto.DocumentUri
type Tree = sitter.Tree
type Node = sitter.Node
type Position = proto.Position
