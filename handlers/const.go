package handlers

import (
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
	proto "github.com/tliron/glsp/protocol_3_16"
	serv "github.com/tliron/glsp/server"
)

var (
	documents map[proto.DocumentUri]*sitter.Tree = make(map[string]*sitter.Tree)
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
