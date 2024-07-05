package handlers

import (
	"context"
	"encoding/json"
	"net/url"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
	serv "github.com/tliron/glsp/server"
)

func CreateServer(handlers glsp.Handler) {
	server = serv.NewServer(handlers, "familymarkup", false)
	server.RunStdio()
}

func logDebug(msg string, data any) {
	if server == nil || server.Log.GetMaxLevel() < 2 {
		return
	}

	str, _ := json.MarshalIndent(data, "", "  ")
	server.Log.Debugf(msg, str)
}

func getTree(uri proto.DocumentUri) (*sitter.Tree, error) {
	tree, ok := documents[uri]

	if ok {
		return tree, nil
	}

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	src, err := os.ReadFile(u.Path)

	if err != nil {
		return nil, err
	}

	tree, err = parser.ParseCtx(context.Background(), nil, src)

	if err != nil {
		return nil, err
	}

	documents[uri] = tree

	return tree, nil
}
