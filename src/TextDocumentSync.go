package src

import (
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocOpen(context *glsp.Context, params *proto.DidOpenTextDocumentParams) error {
	logDebug("DocOpen %s", params)

	_, err := openDocText(params.TextDocument.URI, params.TextDocument.Text, nil)

	return err
}

func DocClose(context *glsp.Context, params *proto.DidCloseTextDocumentParams) error {
	logDebug("DocClose %s", params)

	delete(documents, params.TextDocument.URI)

	return nil
}

func DocChange(ctx *glsp.Context, params *proto.DidChangeTextDocumentParams) error {
	doc, err := openDoc(params.TextDocument.URI)

	if err != nil {
		return err
	}

	for _, wrap := range params.ContentChanges {
		var err error

		switch change := wrap.(type) {
		case proto.TextDocumentContentChangeEventWhole:
			err = doc.SetText(change.Text)

		case proto.TextDocumentContentChangeEvent:
			err = doc.Change(&change)
		}

		setTree(params.TextDocument.URI, doc.Tree)

		if err != nil {
			return err
		}
	}

	root.DirtyUris.Set(params.TextDocument.URI)

	return nil
}
