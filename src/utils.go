package src

import (
	"encoding/json"
	urlParser "net/url"
	"strings"

	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tliron/glsp"
	serv "github.com/tliron/glsp/server"
)

var logOnly string

func CreateServer(handlers glsp.Handler) {
	server = serv.NewServer(handlers, "familymarkup", false)
	server.RunStdio()
}

func createParser() *sitter.Parser {
	p := sitter.NewParser()
	p.SetLanguage(familymarkup.GetLanguage())
	return p
}

func getParser() *sitter.Parser {
	if parser == nil {
		parser = createParser()
	}

	return parser
}

func logDebug(msg string, data any) {
	if logOnly != "" && !strings.HasPrefix(msg, logOnly) {
		return
	}

	if server == nil || server.Log.GetMaxLevel() < 2 {
		return
	}

	str, _ := json.MarshalIndent(data, "", "  ")
	server.Log.Debugf(msg, str)
}

func Debugf(msg string, args ...any) {
	server.Log.Debugf(msg, args...)
}

func LogOnly(prefix string) {
	logOnly = prefix
}

func uriToPath(uri Uri) (string, error) {
	if strings.HasPrefix(uri, "/") {
		return uri, nil
	}

	url, err := urlParser.Parse(uri)

	if err != nil {
		return "", err
	}

	return url.Path, nil
}
