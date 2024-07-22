package handlers

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
	logDebug("DocChange %s", params.TextDocument)

	doc, err := openDoc(params.TextDocument.URI)

	if err != nil {
		return err
	}

	for _, wrap := range params.ContentChanges {
		var err error

		switch change := wrap.(type) {
		case proto.TextDocumentContentChangeEventWhole:
			logDebug("DocChange change whole %s", change)
			err = doc.SetText(change.Text)

		case proto.TextDocumentContentChangeEvent:
			logDebug("DocChange change range %s", change)
			err = doc.Change(&change)

		default:
			logDebug("DocChange unknown type %s", change)
		}

		trees[params.TextDocument.URI] = doc.Tree

		if err != nil {
			logDebug("DocChange err %s", err.Error())
			return err
		}
	}

	logDebug("DocChange %s", "success")

	return nil
}
