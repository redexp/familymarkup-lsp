package main_test

import (
	"github.com/redexp/familymarkup-lsp/state"
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

func TestGetTextByRange(t *testing.T) {
	doc := state.CreateDoc("", "Test\n\nName+Name")

	list := []struct {
		proto.Range
		Text string
	}{
		{
			Range: proto.Range{
				Start: proto.Position{0, 0},
				End:   proto.Position{2, 0},
			},
			Text: "Test\n\n",
		},
		{
			Range: proto.Range{
				Start: proto.Position{0, 0},
				End:   proto.Position{2, 4},
			},
			Text: "Test\n\nName",
		},
		{
			Range: proto.Range{
				Start: proto.Position{0, 2},
				End:   proto.Position{2, 3},
			},
			Text: "st\n\nNam",
		},
		{
			Range: proto.Range{
				Start: proto.Position{2, 1},
				End:   proto.Position{2, 3},
			},
			Text: "am",
		},
		{
			Range: proto.Range{
				Start: proto.Position{2, 1},
				End:   proto.Position{2, 5},
			},
			Text: "ame+",
		},
		{
			Range: proto.Range{
				Start: proto.Position{1, 0},
				End:   proto.Position{2, 4},
			},
			Text: "\nName",
		},
		{
			Range: proto.Range{
				Start: proto.Position{0, 0},
				End:   proto.Position{0, 10},
			},
			Text: "Test\n",
		},
		{
			Range: proto.Range{
				Start: proto.Position{0, 10},
				End:   proto.Position{2, 0},
			},
			Text: "\n",
		},
		{
			Range: proto.Range{
				Start: proto.Position{4, 0},
				End:   proto.Position{5, 0},
			},
			Text: "",
		},
		{
			Range: proto.Range{
				Start: proto.Position{0, 0},
				End:   proto.Position{0, 0},
			},
			Text: "",
		},
	}

	for i, item := range list {
		text := doc.GetTextByRange(item.Range)

		if text != item.Text {
			t.Errorf("%d - got: %s; expect: %s", i+1, text, item.Text)
		}
	}
}
