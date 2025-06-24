//go:build wasm || wasip1

package providers

import (
	"github.com/sourcegraph/jsonrpc2"
	"golang.org/x/net/context"
	"os"
)

func StartServer() {
	stream := &ReadWriteCloser{
		reader: os.Stdin,
		writer: os.Stdout,
	}

	handler := CreateRequestHandler()

	conn := jsonrpc2.NewConn(
		context.Background(),
		jsonrpc2.NewBufferedStream(stream, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.HandlerWithError(handler.RpcHandle),
	)

	<-conn.DisconnectNotify()
}
