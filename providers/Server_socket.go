//go:build !wasm && !wasip1

package providers

import (
	"flag"
	"fmt"
	serv "github.com/tliron/glsp/server"
)

func StartServer() {
	socket := flag.Int("web-socket", 0, "socket number")
	flag.Parse()

	if socket == nil || *socket == 0 {
		panic("--web-socket required")
	}

	server := serv.NewServer(CreateRequestHandler(), "familymarkup", false)
	err := server.RunWebSocket(fmt.Sprintf("127.0.0.1:%d", *socket))

	if err != nil {
		panic(err)
	}
}
