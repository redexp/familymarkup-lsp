package src

import (
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

var docDiagnostic = createDocDebouncer()

func DocOpen(context *glsp.Context, params *proto.DidOpenTextDocumentParams) (err error) {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	doc, err := openDocText(uri, params.TextDocument.Text, nil)

	if err != nil {
		return
	}

	root.DirtyUris.Set(uri)

	PublishDiagnostics(context, uri, doc)

	return
}

func DocClose(context *glsp.Context, params *proto.DidCloseTextDocumentParams) error {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return err
	}

	closeDoc(uri)

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
	docDiagnostic.Set(uri, ctx)

	return nil
}

func DocCreate(ctx *glsp.Context, params *proto.CreateFilesParams) error {
	for _, file := range params.Files {
		err := setDirtyUri(ctx, file.URI)

		if err != nil {
			return err
		}
	}

	return nil
}

func DocRename(ctx *glsp.Context, params *proto.RenameFilesParams) error {
	for _, file := range params.Files {
		err := setDirtyUri(ctx, file.OldURI, file.NewURI)

		if err != nil {
			return err
		}
	}

	return nil
}

func DocDelete(ctx *glsp.Context, params *proto.DeleteFilesParams) error {
	for _, file := range params.Files {
		err := setDirtyUri(ctx, file.URI)

		if err != nil {
			return err
		}
	}

	return nil
}

func setDirtyUri(ctx *glsp.Context, uris ...Uri) error {
	for _, uri := range uris {
		uri, err := normalizeUri(uri)

		if err != nil {
			return err
		}

		root.DirtyUris.Set(uri)
		docDiagnostic.Set(uri, ctx)
	}

	return nil
}
