package main

import (
	lsp "github.com/redexp/familymarkup-lsp/providers"
	"github.com/spf13/pflag"
)

func init() {
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflag.Int("web-socket", 0, "Start websocket server on port")
	pflag.Parse()
}

func main() {
	err := lsp.StartServer()

	if err != nil {
		panic(err)
	}
}
