package providers

import protocol "github.com/tliron/glsp/protocol_3_16"

func NewProtocolHandlers() *protocol.Handler {
	return &protocol.Handler{
		Initialize:                          Initialize,
		Initialized:                         Initialized,
		SetTrace:                            SetTrace,
		CancelRequest:                       CancelRequest,
		TextDocumentSemanticTokensFull:      SemanticTokensFull,
		TextDocumentSemanticTokensFullDelta: SemanticTokensDelta,
		TextDocumentDidOpen:                 DocOpen,
		TextDocumentDidChange:               DocChange,
		TextDocumentDidClose:                DocClose,
		WorkspaceDidCreateFiles:             DocCreate,
		WorkspaceDidRenameFiles:             DocRename,
		WorkspaceDidDeleteFiles:             DocDelete,
		TextDocumentCompletion:              Completion,
		TextDocumentDefinition:              Definition,
		TextDocumentReferences:              References,
		TextDocumentTypeDefinition:          TypeDefinition,
		TextDocumentHover:                   Hover,
		TextDocumentDocumentHighlight:       DocumentHighlight,
		TextDocumentPrepareRename:           PrepareRename,
		TextDocumentRename:                  Rename,
		TextDocumentFoldingRange:            FoldingRange,
		TextDocumentCodeAction:              CodeAction,
		TextDocumentDocumentSymbol:          DocSymbols,
		TextDocumentFormatting:              DocFormating,
		TextDocumentRangeFormatting:         RangeFormating,
		TextDocumentOnTypeFormatting:        LineFormating,
		CodeActionResolve:                   CodeActionResolve,
	}
}

func NewWorkspaceHandlers() *WorkspaceHandler {
	return &WorkspaceHandler{
		WorkspaceSymbol:        AllSymbols,
		WorkspaceSymbolResolve: ResolveSymbol,
	}
}

func NewTreeHandlers() *TreeHandlers {
	return &TreeHandlers{
		TreeFamilies:  TreeFamilies,
		TreeRelations: TreeRelations,
		TreeMembers:   TreeMembers,
		TreeLocation:  TreeLocation,
	}
}

func NewConfigurationHandlers() *ConfigurationHandlers {
	return &ConfigurationHandlers{
		Change: ConfigurationChange,
	}
}
