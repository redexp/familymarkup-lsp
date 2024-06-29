package main_test

import (
	"os"
	"path/filepath"
	"testing"

	main "github.com/redexp/familymarkup-lsp"
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
	_, err := main.Initialize(nil, nil)

	if err != nil {
		t.Errorf("Initialize: %v", err)
	}
}

func TestSemanticTokensFull(t *testing.T) {
	Initialize(t)

	res, err := main.SemanticTokensFull(nil, &proto.SemanticTokensParams{
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

	_, err = main.SemanticTokensFull(nil, &proto.SemanticTokensParams{
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

	ranges := make([]uint32, 0)

	ranges = append(ranges, 1, 9, 12)
	ranges = append(ranges, 1, 3, 5)
	ranges = append(ranges, 3, 3, 4)
	ranges = append(ranges, 3, 4, 6)
	ranges = append(ranges, 3, 5, 6)
	ranges = append(ranges, 3, 6, 10)
	ranges = append(ranges, 5, 9, 5)
	ranges = append(ranges, 7, 9, 1)
	ranges = append(ranges, 7, 10, 2)
	ranges = append(ranges, 10, 10, 1)

	count := len(ranges)

	for i := 0; i < count; i += 3 {
		startLine := ranges[i]
		endLine := ranges[i+1]
		tokensCount := ranges[i+2]

		res, err := main.SemanticTokensRange(nil, &proto.SemanticTokensRangeParams{
			TextDocument: proto.TextDocumentIdentifier{
				URI: "file://" + getTestRoot("semanticTokens.txt", t),
			},
			Range: proto.Range{
				Start: proto.Position{
					Line:      startLine - 1,
					Character: 0,
				},
				End: proto.Position{
					Line:      endLine - 1,
					Character: 100,
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

		if len(data.Data) != int(5*tokensCount) {
			t.Errorf("%d tokens %d when should be %d", i/3, len(data.Data)/5, tokensCount)
		}
	}
}
