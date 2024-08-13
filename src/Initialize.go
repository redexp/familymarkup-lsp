package src

import (
	"slices"
	"strings"

	"github.com/redexp/textdocument"
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Initialize(ctx *glsp.Context, params *proto.InitializeParams) (any, error) {
	logDebug("Initialize %s", params)

	root = createRoot()

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
				Range:  false,
				Legend: *legend,
			},
			CompletionProvider: &proto.CompletionOptions{},
			Workspace: &proto.ServerCapabilitiesWorkspace{
				WorkspaceFolders: &proto.WorkspaceFoldersServerCapabilities{
					Supported: &proto.True,
				},
			},
			DefinitionProvider: true,
			HoverProvider:      true,
			ReferencesProvider: true,
		},
	}

	if params.WorkspaceFolders != nil {
		for _, folder := range params.WorkspaceFolders {
			err := readTreesFromDir(folder.URI, func(tree *Tree, text []byte, path string) error {
				return root.Update(tree, text, path)
			})

			if err != nil {
				return nil, err
			}
		}

		root.UpdateUnknownRefs()
	}

	logDebug("Initialize RESULT %s", res)

	return res, nil
}

func GetLegend() (*proto.SemanticTokensLegend, textdocument.HighlightLegend, error) {
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

	typesMap := make(textdocument.HighlightLegend, len(legend))

	for i, name := range legend {
		parts := strings.Split(name, ".")
		first := parts[0]
		rest := parts[1:]
		tt := textdocument.TokenType{}

		tt.Type = add(&types, first, mapTypes)
		modIndexes := make([]uint32, 0)

		for _, m := range rest {
			index := add(&modifiers, m, mapMod)
			modIndexes = append(modIndexes, index)
		}

		tt.Modifiers = textdocument.BitMask(modIndexes)

		typesMap[i] = tt
	}

	tokensLegend := &proto.SemanticTokensLegend{
		TokenTypes:     types,
		TokenModifiers: modifiers,
	}

	return tokensLegend, typesMap, nil
}
