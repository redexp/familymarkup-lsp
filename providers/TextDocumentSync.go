package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocOpen(_ *Ctx, params *proto.DidOpenTextDocumentParams) (err error) {
	uri := NormalizeUri(params.TextDocument.URI)

	text := params.TextDocument.Text

	if doc, ok := root.Docs[uri]; ok && doc.Text == text {
		doc.Open = true
		return
	}

	root.DirtyUris.SetText(uri, UriOpen, text)

	return
}

func DocClose(_ *Ctx, params *proto.DidCloseTextDocumentParams) (err error) {
	uri := NormalizeUri(params.TextDocument.URI)

	root.CloseDoc(uri)

	return
}

func DocChange(_ *Ctx, params *proto.DidChangeTextDocumentParams) (err error) {
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
				root.DirtyUris.SetText(uri, UriChange, change.Text)
				continue
			}

			root.DirtyUris.ChangeText(doc, change.Range, change.Text)
		}
	}

	return
}

func DocCreate(_ *Ctx, _ *proto.CreateFilesParams) error {
	return nil
}

func DocRename(_ *Ctx, params *proto.RenameFilesParams) error {
	for _, file := range params.Files {
		oldUri := NormalizeUri(file.OldURI)
		newUri := NormalizeUri(file.NewURI)

		doc, ok := root.Docs[oldUri]

		if !ok {
			continue
		}

		root.DirtyUris.Set(oldUri, UriDelete)
		root.DirtyUris.SetText(newUri, UriCreate, doc.Text)
	}

	return nil
}

func DocDelete(_ *Ctx, params *proto.DeleteFilesParams) error {
	for _, file := range params.Files {
		uri := NormalizeUri(file.URI)

		root.DirtyUris.Set(uri, UriDelete)
	}

	return nil
}
