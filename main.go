package main

import (
	"fmt"
	"os"

	lsp "github.com/redexp/familymarkup-lsp/providers"
	"github.com/spf13/pflag"
	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
)

func init() {
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true

	pflag.Int("web-socket", 0, "Start websocket server on port")

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
	server := lsp.CreateServer(
		lsp.NewProtocolHandlers(),
		lsp.NewWorkspaceHandlers(),
		lsp.NewTreeHandlers(),
		lsp.NewConfigurationHandlers(),
	)

	webSocketPort, err := pflag.CommandLine.GetInt("web-socket")

	if webSocketPort > 0 {
		if err != nil {
			panic(err)
		}

		server.RunWebSocket(fmt.Sprintf("127.0.0.1:%d", webSocketPort))
	} else {
		server.RunStdio()
	}
}
