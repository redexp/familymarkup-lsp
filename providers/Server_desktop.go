//go:build linux || windows

package providers

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/tliron/glsp"
	serv "github.com/tliron/glsp/server"
)

func StartServer() error {
	webSocketPort, err := pflag.CommandLine.GetInt("web-socket")

	if err != nil {
		return err
	}

	server := CreateServer(
		NewProtocolHandlers(),
		NewWorkspaceHandlers(),
		NewTreeHandlers(),
		NewConfigurationHandlers(),
	)

	if webSocketPort > 0 {
		return server.RunWebSocket(fmt.Sprintf("127.0.0.1:%d", webSocketPort))
	}

	return server.RunStdio()
}

func CreateServer(handlers ...glsp.Handler) *serv.Server {
	return serv.NewServer(CreateRequestHandler(handlers...), "familymarkup", false)
}
