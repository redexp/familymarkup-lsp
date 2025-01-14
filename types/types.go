package types

import (
	"github.com/redexp/textdocument"
	proto "github.com/tliron/glsp/protocol_3_16"
	sitter "github.com/tree-sitter/go-tree-sitter"
)

type TextDocument = textdocument.TextDocument
type Uri = proto.DocumentUri
type Tree = sitter.Tree
type Node = sitter.Node
type Point = sitter.Point
type QueryCapture = sitter.QueryCapture
type Position = proto.Position
type Range = proto.Range
