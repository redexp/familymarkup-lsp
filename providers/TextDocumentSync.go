package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

var docDiagnostic = createDocDebouncer()

func DocOpen(context *glsp.Context, params *proto.DidOpenTextDocumentParams) (err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	doc, err := OpenDocText(uri, params.TextDocument.Text, GetTree(uri))

	if err != nil {
		return
	}

	PublishDiagnostics(context, uri, doc)

	return
}

func DocClose(context *glsp.Context, params *proto.DidCloseTextDocumentParams) error {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return err
	}

	CloseDoc(uri)

	return nil
}

func DocChange(ctx *glsp.Context, params *proto.DidChangeTextDocumentParams) error {
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

func DocCreate(ctx *glsp.Context, params *proto.CreateFilesParams) error {
	for _, file := range params.Files {
		err := setDirtyUri(ctx, file.URI, FileCreate)

		if err != nil {
			return err
		}
	}

	diagnosticOpenDocs(ctx)

	return nil
}

func DocRename(ctx *glsp.Context, params *proto.RenameFilesParams) error {
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

func DocDelete(ctx *glsp.Context, params *proto.DeleteFilesParams) error {
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

func setDirtyUri(ctx *glsp.Context, uri Uri, state uint8) error {
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

func diagnosticOpenDocs(ctx *glsp.Context) {
	for uri := range GetOpenDocsIter() {
		docDiagnostic.Set(uri, ctx)
	}
}
