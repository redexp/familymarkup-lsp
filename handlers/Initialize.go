package handlers

import (
	"math"
	"slices"
	"strings"

	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Initialize(ctx *glsp.Context, params *proto.InitializeParams) (any, error) {
	// logDebug("Initialize WorkspaceFolders %s", params.WorkspaceFolders)

	parser = createParser()

	legend, types, err := GetLegend()

	if err != nil {
		return nil, err
	}

	typesMap = types
	syncType := proto.TextDocumentSyncKindIncremental

	res := &proto.InitializeResult{
		ServerInfo: &proto.InitializeResultServerInfo{
			Name: "familymarkup",
		},
		Capabilities: proto.ServerCapabilities{
			TextDocumentSync: proto.TextDocumentSyncOptions{
				OpenClose: &proto.True,
				Change:    &syncType,
			},
			SemanticTokensProvider: proto.SemanticTokensOptions{
				Full:   true,
				Range:  true,
				Legend: *legend,
			},
			Workspace: &proto.ServerCapabilitiesWorkspace{
				WorkspaceFolders: &proto.WorkspaceFoldersServerCapabilities{
					Supported: &proto.True,
				},
			},
		},
	}

	// logDebug("Initialize res %s", res)

	return res, nil
}

func GetLegend() (*proto.SemanticTokensLegend, []TokenType, error) {
	legend, err := familymarkup.GetHighlightLegend()

	if err != nil {
		return nil, nil, err
	}

	types := make([]string, 0)
	modifiers := make([]string, 0)

	mapTypes := map[string]string{
		"constant": "variable",
	}

	mapMod := map[string]string{
		"def": "definition",
		"ref": "reference",
	}

	add := func(list *[]string, item string, hash map[string]string) uint32 {
		if v, ok := hash[item]; ok {
			item = v
		}

		if !slices.Contains(*list, item) {
			*list = append(*list, item)
		}

		return uint32(slices.Index(*list, item))
	}

	typesMap := make([]TokenType, len(legend))

	for i, name := range legend {
		parts := strings.Split(name, ".")
		first := parts[0]
		rest := parts[1:]
		tt := TokenType{}

		tt.Type = add(&types, first, mapTypes)

		for _, m := range rest {
			index := add(&modifiers, m, mapMod)
			bit := math.Pow(2, float64(index))
			tt.Mod = tt.Mod | uint32(bit)
		}

		typesMap[i] = tt
	}

	tokensLegend := &proto.SemanticTokensLegend{
		TokenTypes:     types,
		TokenModifiers: modifiers,
	}

	return tokensLegend, typesMap, nil
}
