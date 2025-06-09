package main_test

import (
	"os"
	"path/filepath"
	"testing"

	h "github.com/redexp/familymarkup-lsp/providers"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func getCwd(file string, t *testing.T) string {
	dir, err := os.Getwd()

	if err != nil {
		t.Errorf("Gerwd: %v", err)
	}

	return filepath.Join(dir, file)
}

func getTestRoot(file string, t *testing.T) string {
	return getCwd(filepath.Join("test", "root", file), t)
}

func Initialize(t *testing.T) {
	_, err := h.Initialize(nil, nil)

	if err != nil {
		t.Errorf("Initialize: %v", err)
	}
}

func TestSemanticTokensFull(t *testing.T) {
	Initialize(t)

	res, err := h.SemanticTokensFull(nil, &proto.SemanticTokensParams{
		TextDocument: proto.TextDocumentIdentifier{
			URI: "file://" + getTestRoot("semanticTokens.txt", t),
		},
	})

	if err != nil {
		t.Errorf("SemanticTokensFull: %v", err)
	}

	if res == nil || res.Data == nil {
		t.Errorf("res is nil")
	}

	if len(res.Data) != 5*18 {
		t.Errorf("res.Data len %d expected %d", len(res.Data)/5, 18)
	}

	_, err = h.SemanticTokensFull(nil, &proto.SemanticTokensParams{
		TextDocument: proto.TextDocumentIdentifier{
			URI: "file://" + getTestRoot("not-exist.txt", t),
		},
	})

	if err == nil {
		t.Errorf("should return error")
	}
}
