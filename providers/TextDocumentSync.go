package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocOpen(ctx *Ctx, params *proto.DidOpenTextDocumentParams) (err error) {
	uri := NormalizeUri(params.TextDocument.URI)

	text := params.TextDocument.Text

	if doc, ok := root.Docs[uri]; ok && doc.Text == text {
		doc.Open = true
		return
	}

	root.DirtyUris.SetText(uri, UriOpen, text)

	scheduleDiagnostic(ctx, uri)

	return
}

func DocClose(_ *Ctx, params *proto.DidCloseTextDocumentParams) (err error) {
	uri := NormalizeUri(params.TextDocument.URI)

	root.CloseDoc(uri)

	return
}

func DocChange(ctx *Ctx, params *proto.DidChangeTextDocumentParams) (err error) {
	uri := NormalizeUri(params.TextDocument.URI)

	for _, wrap := range params.ContentChanges {
		switch change := wrap.(type) {
		case proto.TextDocumentContentChangeEventWhole:
			root.DirtyUris.SetText(uri, UriChange, change.Text)

		case proto.TextDocumentContentChangeEvent:
			if change.Range == nil {
				root.DirtyUris.SetText(uri, UriChange, change.Text)
				continue
			}

			doc, ok := root.Docs[uri]

			if !ok {
				continue
			}

			root.DirtyUris.ChangeText(doc, change.Range, change.Text)
		}
	}

	scheduleDiagnostic(ctx, uri)

	return
}

func DocRename(ctx *Ctx, params *proto.RenameFilesParams) error {
	for _, file := range params.Files {
		oldUri := NormalizeUri(file.OldURI)
		newUri := NormalizeUri(file.NewURI)

		doc, ok := root.Docs[oldUri]

		if !ok {
			continue
		}

		root.DirtyUris.Set(oldUri, UriDelete)
		root.DirtyUris.SetText(newUri, UriCreate, doc.Text)

		scheduleDiagnostic(ctx, newUri)
	}

	return nil
}

func DocDelete(ctx *Ctx, params *proto.DeleteFilesParams) error {
	for _, file := range params.Files {
		uri := NormalizeUri(file.URI)

		root.DirtyUris.Set(uri, UriDelete)

		scheduleDiagnostic(ctx, uri)
	}

	return nil
}
