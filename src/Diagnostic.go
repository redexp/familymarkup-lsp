package src

import (
	"time"

	"github.com/bep/debounce"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

type DocDebouncer struct {
	Docs     map[Uri]*glsp.Context
	Debounce func(func())
}

func PublishDiagnostics(ctx *glsp.Context, uri Uri, doc *TextDocument) {
	if !supportDiagnostics {
		return
	}

	var err error

	if doc == nil {
		doc, err = openDoc(uri)

		if err != nil {
			logDebug("Diagnostic open doc error: %s", err.Error())
			return
		}
	}

	list := make([]proto.Diagnostic, 0)

	for _, ref := range root.UnknownRefs {
		if ref.Uri != uri {
			continue
		}

		node := ref.Node
		message := "Unknown person"

		if node.Type() == "name_ref" {
			f := root.FindFamily(ref.Surname)

			if f != nil {
				node = node.NamedChild(1)
			} else {
				node = node.NamedChild(0)
				message = "Unknown family"
			}
		}

		r, err := doc.NodeToRange(node)

		if err != nil {
			logDebug("Diagnostic error: %s", err.Error())
			continue
		}

		list = append(list, proto.Diagnostic{
			Range:   *r,
			Message: message,
		})
	}

	ctx.Notify(proto.ServerTextDocumentPublishDiagnostics, proto.PublishDiagnosticsParams{
		URI: uri,
		// TODO add version
		Diagnostics: list,
	})
}

func createDocDebouncer() *DocDebouncer {
	return &DocDebouncer{
		Docs:     make(map[string]*glsp.Context),
		Debounce: debounce.New(200 * time.Millisecond),
	}
}

func (dd *DocDebouncer) Set(uri Uri, ctx *glsp.Context) {
	dd.Docs[uri] = ctx
	dd.Debounce(func() {
		dd.Flush()
	})
}

func (dd *DocDebouncer) Flush() {
	root.UpdateDirty()

	for uri, ctx := range dd.Docs {
		PublishDiagnostics(ctx, uri, nil)
		delete(dd.Docs, uri)
	}
}
