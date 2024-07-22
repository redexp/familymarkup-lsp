package handlers

import (
	"github.com/redexp/textdocument"
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
	proto "github.com/tliron/glsp/protocol_3_16"
	serv "github.com/tliron/glsp/server"
)

type TextDocument = textdocument.TextDocument
type Uri = proto.DocumentUri
type Tree = sitter.Tree

var (
	documents map[Uri]*TextDocument = make(map[Uri]*TextDocument)
	trees     map[Uri]*Tree         = make(map[Uri]*Tree)
	parser    *sitter.Parser
	typesMap  []TokenType
	server    *serv.Server
)

type TokenType struct {
	Type uint32
	Mod  uint32
}

func createParser() *sitter.Parser {
	p := sitter.NewParser()
	p.SetLanguage(familymarkup.GetLanguage())
	return p
}
