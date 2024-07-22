package main

import (
	h "github.com/redexp/familymarkup-lsp/handlers"
	"github.com/spf13/pflag"
	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func init() {
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true

	logLevel := pflag.IntP("log-level", "l", -4, "log level: -4 - None (Default), -3 - Critical, -2 - Error, -1 - Warning, 0 - Notice, 1 - Info, 2 - Debug")
	logFile := pflag.StringP("log-file", "f", "", "path to log file")

	pflag.Parse()

	if *logFile == "" {
		logFile = nil
	}

	commonlog.Configure(*logLevel, logFile)
}

func main() {
	h.CreateServer(&proto.Handler{
		Initialize:                      h.Initialize,
		TextDocumentSemanticTokensFull:  h.SemanticTokensFull,
		TextDocumentSemanticTokensRange: h.SemanticTokensRange,
		TextDocumentDidOpen:             h.DocOpen,
		TextDocumentDidChange:           h.DocChange,
		TextDocumentDidClose:            h.DocClose,
	})
}
