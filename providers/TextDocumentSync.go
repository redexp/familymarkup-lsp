package providers

import (
	. "github.com/redexp/familymarkup-lsp/state"
	"github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
	"strings"
	"sync"
)

func DocOpen(_ *Ctx, params *proto.DidOpenTextDocumentParams) (err error) {
	uri := NormalizeUri(params.TextDocument.URI)

	if !utils.IsFamilyUri(uri) {
		return
	}

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

		if ok {
			root.DirtyUris.Set(oldUri, UriDelete)
			root.DirtyUris.SetText(newUri, UriCreate, doc.Text)
			continue
		}

		if utils.IsMarkdownUri(oldUri) {
			root.DirtyUris.Set(oldUri, UriDelete)
			root.DirtyUris.Set(newUri, UriCreate)
			continue
		}

		if utils.IsFamilyUri(oldUri) {
			continue
		}

		oldFolder := toFolderUri(oldUri)
		newFolder := toFolderUri(newUri)

		var wg sync.WaitGroup

		wg.Add(3)

		go func() {
			defer wg.Done()

			for uri, doc := range root.Docs {
				if !strings.HasPrefix(uri, oldFolder) {
					continue
				}

				root.DirtyUris.Set(uri, UriDelete)

				newUri = strings.Replace(uri, oldFolder, newFolder, 1)

				if _, ok := root.Docs[newUri]; ok {
					continue
				}

				root.DirtyUris.SetText(newUri, UriCreate, doc.Text)
			}
		}()

		go func() {
			defer wg.Done()

			for uri, item := range root.UnknownFiles {
				if !strings.HasPrefix(uri, oldFolder) {
					continue
				}

				delete(root.UnknownFiles, uri)

				newUri = strings.Replace(uri, oldFolder, newFolder, 1)

				root.UnknownFiles[newUri] = item
			}
		}()

		go func() {
			defer wg.Done()

			for mem := range root.MembersIter() {
				if strings.HasPrefix(mem.InfoUri, oldFolder) {
					mem.InfoUri = strings.Replace(mem.InfoUri, oldFolder, newFolder, 1)
				}
			}
		}()

		wg.Wait()
	}

	return nil
}

func DocDelete(_ *Ctx, params *proto.DeleteFilesParams) error {
	for _, file := range params.Files {
		uri := NormalizeUri(file.URI)

		if _, ok := root.Docs[uri]; ok {
			root.DirtyUris.Set(uri, UriDelete)
			continue
		}

		if utils.IsMarkdownUri(uri) {
			root.DirtyUris.Set(uri, UriDelete)
			continue
		}

		if utils.IsFamilyUri(uri) {
			continue
		}

		folder := toFolderUri(uri)

		var wg sync.WaitGroup

		wg.Add(3)

		go func() {
			defer wg.Done()

			for uri := range root.Docs {
				if strings.HasPrefix(uri, folder) {
					root.DirtyUris.Set(uri, UriDelete)
				}
			}
		}()

		go func() {
			defer wg.Done()

			for uri := range root.UnknownFiles {
				if strings.HasPrefix(uri, folder) {
					delete(root.UnknownFiles, uri)
				}
			}
		}()

		go func() {
			defer wg.Done()

			for mem := range root.MembersIter() {
				if strings.HasPrefix(mem.InfoUri, folder) {
					mem.InfoUri = ""
				}
			}
		}()

		wg.Wait()
	}

	return nil
}

func toFolderUri(uri string) string {
	if !strings.HasSuffix(uri, "/") {
		uri += "/"
	}

	return uri
}
