package providers

import (
	"fmt"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
	"golang.org/x/net/context"
)

func CreateRequestHandler() *RequestHandler {
	return &RequestHandler{
		Handlers: []glsp.Handler{
			NewProtocolHandlers(),
			&WorkspaceHandler{
				WorkspaceSymbol:        AllSymbols,
				WorkspaceSymbolResolve: ResolveSymbol,
			},
			&TreeHandlers{
				TreeFamilies:  TreeFamilies,
				TreeRelations: TreeRelations,
				TreeMembers:   TreeMembers,
			},
			&ConfigurationHandlers{
				Change: ConfigurationChange,
			},
			&DiagnosticHandler{
				TextDocumentDiagnostic: TextDocumentDiagnostic,
				WorkspaceDiagnostic:    WorkspaceDiagnostic,
			},
			&SvgHandlers{
				Document: SvgDocument,
			},
		},
	}
}

func NewProtocolHandlers() *proto.Handler {
	return &proto.Handler{
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

type RequestHandler struct {
	Handlers []glsp.Handler
}

func (req *RequestHandler) RpcHandle(c context.Context, conn *jsonrpc2.Conn, r *jsonrpc2.Request) (res any, err error) {
	if r.Method == "exit" {
		err = conn.Close()
		return nil, err
	}

	ctx := &glsp.Context{
		Method: r.Method,
		Notify: func(method string, params any) {
			_ = conn.Notify(c, method, params)
		},
	}

	if r.Params != nil {
		ctx.Params = *r.Params
	}

	var validMethod bool
	var validParams bool

	res, validMethod, validParams, err = req.Handle(ctx)

	if !validMethod {
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeMethodNotFound,
			Message: fmt.Sprintf("Method not found: %s", r.Method),
		}
	}

	if !validParams {
		e := &jsonrpc2.Error{
			Code: jsonrpc2.CodeInvalidParams,
		}

		if err != nil {
			e.Message = err.Error()
		}

		err = e
	}

	return res, err
}

func (req *RequestHandler) Handle(ctx *Ctx) (res any, validMethod bool, validParams bool, err error) {
	for _, h := range req.Handlers {
		res, validMethod, validParams, err = h.Handle(ctx)

		if validMethod {
			return
		}
	}

	return
}
