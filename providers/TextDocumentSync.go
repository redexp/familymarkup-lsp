package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocOpen(ctx *Ctx, params *proto.DidOpenTextDocumentParams) (err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	root.OpenDocText(uri, params.TextDocument.Text)

	scheduleDiagnostic(ctx, uri)

	return
}

func DocClose(_ *Ctx, params *proto.DidCloseTextDocumentParams) error {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return err
	}

	root.CloseDoc(uri)

	return nil
}

func DocChange(ctx *Ctx, params *proto.DidChangeTextDocumentParams) error {
	root.UpdateLock.Lock()
	defer root.UpdateLock.Unlock()

	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return err
	}

	doc, err := root.OpenDoc(uri)

	if err != nil {
		return err
	}

	for _, wrap := range params.ContentChanges {
		switch change := wrap.(type) {
		case proto.TextDocumentContentChangeEventWhole:
			doc.SetText(change.Text)

		case proto.TextDocumentContentChangeEvent:
			doc.Change(change)
		}
	}

	return setDirtyUri(ctx, uri, FileChange)
}

func DocCreate(ctx *Ctx, params *proto.CreateFilesParams) error {
	root.UpdateLock.Lock()
	defer root.UpdateLock.Unlock()

	for _, file := range params.Files {
		err := setDirtyUri(ctx, file.URI, FileCreate)

		if err != nil {
			return err
		}
	}

	diagnosticOpenDocs(ctx)

	return nil
}

func DocRename(ctx *Ctx, params *proto.RenameFilesParams) error {
	root.UpdateLock.Lock()
	defer root.UpdateLock.Unlock()

	for _, file := range params.Files {
		oldUri, err := NormalizeUri(file.OldURI)

		if err != nil {
			return err
		}

		newUri, err := NormalizeUri(file.NewURI)

		if err != nil {
			return err
		}

		doc, ok := root.Docs[oldUri]

		if ok {
			newDoc := *doc
			newDoc.Uri = newUri
			root.Docs[newUri] = &newDoc
		}

		err = setDirtyUri(ctx, file.OldURI, FileDelete)

		if err != nil {
			return err
		}

		err = setDirtyUri(ctx, file.NewURI, FileCreate)

		if err != nil {
			return err
		}
	}

	diagnosticOpenDocs(ctx)

	return nil
}

func DocDelete(ctx *Ctx, params *proto.DeleteFilesParams) (err error) {
	root.UpdateLock.Lock()
	defer root.UpdateLock.Unlock()

	for _, file := range params.Files {
		err = setDirtyUri(ctx, file.URI, FileDelete)

		if err != nil {
			return
		}
	}

	diagnosticOpenDocs(ctx)

	return
}

func setDirtyUri(ctx *Ctx, uri Uri, state uint8) error {
	uri, err := NormalizeUri(uri)

	if err != nil {
		return err
	}

	if IsFamilyUri(uri) || IsMarkdownUri(uri) {
		root.DirtyUris.SetState(uri, state)
		scheduleDiagnostic(ctx, uri)
	}

	return nil
}
