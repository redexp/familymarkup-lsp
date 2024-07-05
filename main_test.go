package main_test

import (
	"os"
	"path/filepath"
	"testing"

	h "github.com/redexp/familymarkup-lsp/handlers"
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

func TestSemanticTokensRange(t *testing.T) {
	Initialize(t)

	type Range struct {
		startLine uint32
		endLine   uint32
		count     uint32
		endChar   uint32
	}

	ranges := []Range{
		{1, 9, 12, 100},
		{1, 3, 5, 100},
		{3, 3, 4, 100},
		{3, 4, 6, 100},
		{3, 5, 6, 100},
		{3, 6, 10, 100},
		{5, 9, 5, 100},
		{7, 9, 1, 100},
		{7, 10, 2, 100},
		{10, 10, 1, 100},
		{1, 3, 3, 6},
	}

	for i, r := range ranges {
		res, err := h.SemanticTokensRange(nil, &proto.SemanticTokensRangeParams{
			TextDocument: proto.TextDocumentIdentifier{
				URI: "file://" + getTestRoot("semanticTokens.txt", t),
			},
			Range: proto.Range{
				Start: proto.Position{
					Line:      r.startLine - 1,
					Character: 0,
				},
				End: proto.Position{
					Line:      r.endLine - 1,
					Character: r.endChar,
				},
			},
		})

		if err != nil {
			t.Errorf("SemanticTokensRange: %v", err)
		}

		if res == nil {
			t.Errorf("res is nil")
		}

		data, ok := res.(*proto.SemanticTokens)

		if !ok {
			t.Errorf("res is not *SemanticTokesn")
		}

		if len(data.Data) != int(5*r.count) {
			t.Errorf("%d tokens %d when should be %d", i, len(data.Data)/5, r.count)
		}
	}
}
