package main_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	h "github.com/redexp/familymarkup-lsp/handlers"
	"github.com/redexp/textdocument"
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	sitter "github.com/smacker/go-tree-sitter"
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

func TestXxx(t *testing.T) {
	text := "Fam\n\nNam + Nas" // 3 _ 1 _ 3
	p := sitter.NewParser()
	p.SetLanguage(familymarkup.GetLanguage())

	doc := textdocument.NewTextDocument(text)
	doc.SetParser(p)

	check := func() error {
		root := doc.Tree.RootNode()
		fmt.Println(doc.Text)
		fmt.Println(root.String())
		caps, err := h.GetCaptures(root)

		if err != nil {
			return err
		}

		for i, cap := range caps {
			node := cap.Node
			fmt.Printf("%d %s %v\n", i, node.String(), node.Range())
		}

		return nil
	}

	err := check()

	if err != nil {
		t.Error(err)
		return
	}

	err = doc.Change(&textdocument.ChangeEvent{
		Range: &proto.Range{
			Start: proto.Position{
				Line:      2,
				Character: 9,
			},
			End: proto.Position{
				Line:      2,
				Character: 9,
			},
		},
		Text: "d",
	})

	if err != nil {
		t.Error(err)
		return
	}

	err = check()

	if err != nil {
		t.Error(err)
		return
	}
}
