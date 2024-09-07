package main

import (
	"os"

	lsp "github.com/redexp/familymarkup-lsp/src"
	"github.com/spf13/pflag"
	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func init() {
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true

	logLevel := pflag.IntP("log-level", "l", -4, "log level: -4 - None (Default), -3 - Critical, -2 - Error, -1 - Warning, 0 - Notice, 1 - Info, 2 - Debug")
	logFile := pflag.StringP("log-file", "f", "", "path to log file")
	logOnly := pflag.StringP("log-only", "p", "", "log only with prefix")
	logClear := pflag.BoolP("log-clear", "c", false, "clear log on start")

	pflag.Parse()

	if *logFile == "" {
		logFile = nil
	}

	lsp.LogOnly(*logOnly)

	if *logClear && logFile != nil {
		os.Truncate(*logFile, 0)
	}

	commonlog.Configure(*logLevel, logFile)
}

func main() {
	protocolHandlers := &proto.Handler{
		Initialize:                     lsp.Initialize,
		Initialized:                    lsp.Initialized,
		SetTrace:                       lsp.SetTrace,
		CancelRequest:                  lsp.CancelRequest,
		TextDocumentSemanticTokensFull: lsp.SemanticTokensFull,
		TextDocumentDidOpen:            lsp.DocOpen,
		TextDocumentDidChange:          lsp.DocChange,
		TextDocumentDidClose:           lsp.DocClose,
		WorkspaceDidCreateFiles:        lsp.DocCreate,
		WorkspaceDidRenameFiles:        lsp.DocRename,
		WorkspaceDidDeleteFiles:        lsp.DocDelete,
		TextDocumentCompletion:         lsp.Completion,
		TextDocumentDefinition:         lsp.Definition,
		TextDocumentHover:              lsp.Hover,
		TextDocumentReferences:         lsp.References,
		TextDocumentDocumentHighlight:  lsp.DocumentHighlight,
		TextDocumentPrepareRename:      lsp.PrepareRename,
		TextDocumentRename:             lsp.Rename,
		TextDocumentFoldingRange:       lsp.FoldingRange,
		TextDocumentCodeAction:         lsp.CodeAction,
		TextDocumentDocumentSymbol:     lsp.Symbol,
		CodeActionResolve:              lsp.CodeActionResolve,
	}

	customHandlers := &lsp.CustomHandlers{
		TreeFamilies:  lsp.TreeFamilies,
		TreeRelations: lsp.TreeRelations,
		TreeMembers:   lsp.TreeMembers,
	}

	lsp.CreateServer(protocolHandlers, customHandlers)
}
