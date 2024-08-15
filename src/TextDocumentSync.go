package src

import (
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocOpen(context *glsp.Context, params *proto.DidOpenTextDocumentParams) error {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return err
	}

	_, err = openDocText(uri, params.TextDocument.Text, nil)

	return err
}

func DocClose(context *glsp.Context, params *proto.DidCloseTextDocumentParams) error {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return err
	}

	delete(documents, uri)

	return nil
}

func DocChange(ctx *glsp.Context, params *proto.DidChangeTextDocumentParams) error {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return err
	}

	doc, err := openDoc(uri)

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

		setTree(uri, doc.Tree)

		if err != nil {
			return err
		}
	}

	root.DirtyUris.Set(uri)

	return nil
}
