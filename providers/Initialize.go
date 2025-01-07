package providers

import (
	"fmt"
	"slices"
	"strings"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	"github.com/redexp/textdocument"
	familymarkup "github.com/redexp/tree-sitter-familymarkup"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Initialize(ctx *Ctx, params *proto.InitializeParams) (any, error) {
	root = CreateRoot(Debugf)

	options, err := GetClientConfiguration(params.InitializationOptions)

	if err == nil {
		SetLocale(options.Locale)
		warnChildrenWithoutRelations = options.WarnChildrenWithoutRelations
	}

	legend, types, err := GetLegend()

	if err != nil {
		return nil, err
	}

	typesMap = types
	syncType := proto.TextDocumentSyncKindIncremental
	fileFilters := proto.FileOperationRegistrationOptions{
		Filters: []proto.FileOperationFilter{
			{
				Scheme: P("file"),
				Pattern: proto.FileOperationPattern{
					Glob: fmt.Sprintf("**/*.{%s}", strings.Join(AllExt, ",")),
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
				Full: proto.SemanticDelta{
					Delta: &proto.True,
				},
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
			ReferencesProvider:        true,
			TypeDefinitionProvider:    true,
			HoverProvider:             true,
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
			DocumentFormattingProvider:      true,
			DocumentRangeFormattingProvider: true,
			DocumentOnTypeFormattingProvider: &proto.DocumentOnTypeFormattingOptions{
				FirstTriggerCharacter: " ",
				MoreTriggerCharacter:  []string{"(", ")", "\n"},
			},
		},
	}

	if params.WorkspaceFolders != nil {
		folders := make([]string, len(params.WorkspaceFolders))

		for i, folder := range params.WorkspaceFolders {
			folders[i] = folder.URI
		}

		err = root.SetFolders(folders)

		if err != nil {
			return nil, err
		}
	}

	supportDiagnostics = params.Capabilities.TextDocument != nil && params.Capabilities.TextDocument.PublishDiagnostics != nil

	return res, nil
}

func Initialized(ctx *Ctx, params *proto.InitializedParams) error {
	diagnosticAllDocs(ctx)

	return nil
}

func SetTrace(ctx *Ctx, params *proto.SetTraceParams) error {
	return nil
}

func CancelRequest(ctx *Ctx, params *proto.CancelParams) error {
	return nil
}

func GetLegend() (*proto.SemanticTokensLegend, textdocument.HighlightLegend, error) {
	legend, err := familymarkup.GetHighlightLegend()

	if err != nil {
		return nil, nil, err
	}

	types := make([]string, 0)
	modifiers := make([]string, 0)

	mapTypes := map[string]string{}

	mapMod := map[string]string{}

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
