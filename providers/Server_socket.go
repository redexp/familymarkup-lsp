//go:build !wasm && !wasip1

package providers

import (
	"flag"
	"fmt"
	serv "github.com/tliron/glsp/server"
)

func StartServer() {
	handler := CreateRequestHandler(
		NewProtocolHandlers(),
		NewWorkspaceHandlers(),
		NewTreeHandlers(),
		NewConfigurationHandlers(),
	)

	server := serv.NewServer(handler, "familymarkup", false)

	socket := flag.Int("web-socket", 0, "socket number")
	flag.Parse()

	if socket == nil || *socket == 0 {
		panic("--web-socket required")
	}

	err := server.RunWebSocket(fmt.Sprintf("127.0.0.1:%d", *socket))

	if err != nil {
		panic(err)
	}
}
