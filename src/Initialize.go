package src

import (
	"slices"
	"strings"
	"sync"

	"github.com/redexp/textdocument"
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Initialize(ctx *glsp.Context, params *proto.InitializeParams) (any, error) {
	root = createRoot()

	legend, types, err := GetLegend()

	if err != nil {
		return nil, err
	}

	typesMap = types
	syncType := proto.TextDocumentSyncKindIncremental
	fileFilters := proto.FileOperationRegistrationOptions{
		Filters: []proto.FileOperationFilter{
			{
				Scheme: pt("file"),
				Pattern: proto.FileOperationPattern{
					Glob: "**/*.{fm,fml,family}",
				},
			},
		},
	}

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
				FileOperations: &proto.ServerCapabilitiesWorkspaceFileOperations{
					DidCreate: &fileFilters,
					DidRename: &fileFilters,
					DidDelete: &fileFilters,
				},
			},
			DefinitionProvider:        true,
			HoverProvider:             true,
			ReferencesProvider:        true,
			DocumentHighlightProvider: true,
			FoldingRangeProvider:      true,
			DocumentSymbolProvider:    true,
			WorkspaceSymbolProvider: WorkspaceSymbolOptions{
				ResolveProvider: true,
			},
			RenameProvider: proto.RenameOptions{
				PrepareProvider: &proto.True,
			},
			CodeActionProvider: proto.CodeActionOptions{
				CodeActionKinds: []proto.CodeActionKind{
					proto.CodeActionKindQuickFix,
				},
				ResolveProvider: &proto.True,
			},
		},
	}

	if params.WorkspaceFolders != nil {
		lock := sync.Mutex{}

		for _, folder := range params.WorkspaceFolders {
			path, err := uriToPath(folder.URI)

			if err != nil {
				return nil, err
			}

			err = readTreesFromDir(path, func(tree *Tree, text []byte, path string) error {
				lock.Lock()
				defer lock.Unlock()

				return root.Update(tree, text, toUri(path))
			})

			if err != nil {
				return nil, err
			}
		}
	}

	supportDiagnostics = params.Capabilities.TextDocument != nil && params.Capabilities.TextDocument.PublishDiagnostics != nil

	return res, nil
}

func Initialized(context *glsp.Context, params *proto.InitializedParams) error {
	waitTreesReady()

	root.UpdateUnknownRefs()

	return nil
}

func SetTrace(context *glsp.Context, params *proto.SetTraceParams) error {
	Debugf("SetTrace: %v", params.Value)
	return nil
}

func CancelRequest(context *glsp.Context, params *proto.CancelParams) error {
	Debugf("CancelRequest: %v", params.ID)
	return nil
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
