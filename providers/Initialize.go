package providers

import (
	"fmt"
	"strings"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Initialize(_ *Ctx, params *proto.InitializeParams) (any, error) {
	root = CreateRoot()

	options, err := GetClientConfiguration(params.InitializationOptions)

	if err == nil {
		err = SetLocale(options.Locale)

		if err != nil {
			return nil, err
		}

		warnChildrenWithoutRelations = options.WarnChildrenWithoutRelations
	}

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

	type obj map[string]any

	res := obj{
		"serverInfo": obj{
			"name": "familymarkup",
		},
		"capabilities": obj{
			"textDocumentSync": obj{
				"openClose": true,
				"change":    proto.TextDocumentSyncKindIncremental,
			},
			"semanticTokensProvider": obj{
				"full": obj{
					"delta": true,
				},
				"range": false,
				"legend": obj{
					"tokenTypes":     Legend.Types,
					"tokenModifiers": Legend.Modifiers,
				},
			},
			"completionProvider": obj{},
			"workspace": obj{
				"workspaceFolders": obj{
					"supported": true,
				},
				"fileOperations": obj{
					"didCreate": fileFilters,
					"didRename": fileFilters,
					"didDelete": fileFilters,
				},
			},
			"definitionProvider":        true,
			"referencesProvider":        true,
			"typeDefinitionProvider":    true,
			"hoverProvider":             true,
			"documentHighlightProvider": true,
			"foldingRangeProvider":      true,
			"documentSymbolProvider":    true,
			"workspaceSymbolProvider": obj{
				"resolveProvider": true,
			},
			"renameProvider": obj{
				"prepareProvider": true,
			},
			"codeActionProvider": obj{
				"codeActionKinds": []proto.CodeActionKind{
					proto.CodeActionKindQuickFix,
				},
				"resolveProvider": true,
			},
			"documentFormattingProvider":      true,
			"documentRangeFormattingProvider": true,
			"documentOnTypeFormattingProvider": proto.DocumentOnTypeFormattingOptions{
				FirstTriggerCharacter: " ",
				MoreTriggerCharacter:  []string{"(", ")", "\n"},
			},
			"diagnosticProvider": obj{
				"interFileDependencies": true,
				"workspaceDiagnostics":  true,
			},
		},
	}

	if params.WorkspaceFolders != nil {
		folders := make([]string, len(params.WorkspaceFolders))

		for i, folder := range params.WorkspaceFolders {
			folders[i] = NormalizeUri(folder.URI)
		}

		root.SetFolders(folders)
	}

	supportDiagnostics = params.Capabilities.TextDocument != nil && params.Capabilities.TextDocument.PublishDiagnostics != nil

	return res, nil
}

func Initialized(_ *Ctx, _ *proto.InitializedParams) (err error) {
	err = root.UpdateDirty()

	return
}

func SetTrace(_ *Ctx, _ *proto.SetTraceParams) error {
	return nil
}

func CancelRequest(_ *Ctx, _ *proto.CancelParams) error {
	return nil
}
