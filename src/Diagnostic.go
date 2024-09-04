package src

import (
	"fmt"
	"time"

	"github.com/bep/debounce"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

type DocDebouncer struct {
	Docs     map[Uri]*glsp.Context
	Debounce func(func())
}

const (
	UnknownFamilyError = iota
	NameDuplicateWarning
)

type DiagnosticData struct {
	Type    uint8  `json:"type"`
	Surname string `json:"surname"`
	Name    string `json:"name"`
}

func PublishDiagnostics(ctx *glsp.Context, uri Uri, doc *TextDocument) {
	if !supportDiagnostics {
		return
	}

	var err error

	if doc == nil {
		doc, err = tempDoc(uri)

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

		if isNameRef(node) {
			f := root.FindFamily(ref.Surname)

			if f != nil {
				node = node.NamedChild(1)
			} else {
				node = node.NamedChild(0)
				message = "Unknown family"
			}
		} else if isNewSurname(node.Parent()) {
			message = "Unknown family"
		}

		r, err := doc.NodeToRange(node)

		if err != nil {
			logDebug("Diagnostic error: %s", err.Error())
			continue
		}

		list = append(list, proto.Diagnostic{
			Severity: pt(proto.DiagnosticSeverityError),
			Range:    *r,
			Message:  message,
			Data: DiagnosticData{
				Type: UnknownFamilyError,
			},
		})
	}

	tempDocs := make(Docs)
	tempDocs[uri] = doc

	for family := range root.FamilyIter() {
		for name, dups := range family.Duplicates {
			member := family.Members[name]
			dups = append(dups, &Duplicate{Member: member})
			count := len(dups)

			var locations []proto.DiagnosticRelatedInformation

			for _, ref := range member.Refs {
				if ref.Uri != uri {
					continue
				}

				r, err := doc.NodeToRange(nameRefName(ref.Node))

				if err != nil {
					logDebug("Diagnostic error: %s", err.Error())
					continue
				}

				if locations == nil {
					d, err := tempDocs.Get(family.Uri)

					if err != nil {
						logDebug("Diagnostic error: %s", err.Error())
						continue
					}

					locations = make([]proto.DiagnosticRelatedInformation, count)
					for i, dup := range dups {
						node := dup.Member.Node
						rr, err := d.NodeToRange(node)

						if err != nil {
							logDebug("Diagnostic error: %s", err.Error())
							continue
						}

						sources := getClosestSources(node)

						locations[i] = proto.DiagnosticRelatedInformation{
							Location: proto.Location{
								URI:   family.Uri,
								Range: *rr,
							},
							Message: fmt.Sprintf("Child of %s", toString(sources, d)),
						}
					}
				}

				list = append(list, proto.Diagnostic{
					Severity:           pt(proto.DiagnosticSeverityWarning),
					Range:              *r,
					Message:            fmt.Sprintf("An unobvious name. There are %d persons with the name %s. Add uniq name alias to one of them", count, name),
					RelatedInformation: locations,
					Data: DiagnosticData{
						Type:    NameDuplicateWarning,
						Surname: family.Name,
						Name:    name,
					},
				})
			}
		}
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
		delete(dd.Docs, uri)

		if !docExist(uri) {
			continue
		}

		PublishDiagnostics(ctx, uri, nil)
	}
}
