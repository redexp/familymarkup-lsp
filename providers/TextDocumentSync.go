package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

var docDiagnostic = createDocDebouncer()

func DocOpen(ctx *Ctx, params *proto.DidOpenTextDocumentParams) (err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	doc, err := OpenDocText(uri, params.TextDocument.Text, GetTree(uri))

	if err != nil {
		return
	}

	PublishDiagnostics(ctx, uri, doc)

	return
}

func DocClose(ctx *Ctx, params *proto.DidCloseTextDocumentParams) error {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return err
	}

	CloseDoc(uri)

	return nil
}

func DocChange(ctx *Ctx, params *proto.DidChangeTextDocumentParams) error {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return err
	}

	doc, err := OpenDoc(uri)

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

		SetTree(uri, doc.Tree)

		if err != nil {
			return err
		}
	}

	return setDirtyUri(ctx, uri, FileChange)
}

func DocCreate(ctx *Ctx, params *proto.CreateFilesParams) error {
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
	for _, file := range params.Files {
		err := RemoveDoc(file.OldURI)

		if err != nil {
			return err
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

func DocDelete(ctx *Ctx, params *proto.DeleteFilesParams) error {
	for _, file := range params.Files {
		err := RemoveDoc(file.URI)

		if err != nil {
			return err
		}

		err = setDirtyUri(ctx, file.URI, FileDelete)

		if err != nil {
			return err
		}
	}

	diagnosticOpenDocs(ctx)

	return nil
}

func setDirtyUri(ctx *Ctx, uri Uri, state uint8) error {
	uri, err := NormalizeUri(uri)

	if err != nil {
		return err
	}

	if IsFamilyUri(uri) || IsMarkdownUri(uri) {
		root.DirtyUris.SetState(uri, state)
		docDiagnostic.Set(uri, ctx)
	}

	return nil
}

func diagnosticOpenDocs(ctx *Ctx) {
	for uri := range GetOpenDocsIter() {
		docDiagnostic.Set(uri, ctx)
	}
}
