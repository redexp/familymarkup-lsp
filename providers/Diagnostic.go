package providers

import (
	"time"

	"github.com/bep/debounce"
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

type DocDebouncer struct {
	Docs     map[Uri]*Ctx
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

func PublishDiagnostics(ctx *Ctx, uri Uri, doc *TextDocument) {
	if !supportDiagnostics {
		return
	}

	var err error

	if doc == nil {
		doc, err = TempDoc(uri)

		if err != nil {
			LogDebug("Diagnostic open doc error: %s", err.Error())
			return
		}
	}

	list := make([]proto.Diagnostic, 0)

	for node := range GetErrorNodesIter(doc.Tree.RootNode()) {
		r, err := doc.NodeToRange(node)

		if err != nil {
			LogDebug("Diagnostic error: %s", err.Error())
			return
		}

		list = append(list, proto.Diagnostic{
			Severity: P(proto.DiagnosticSeverityError),
			Range:    *r,
			Message:  L("syntax_error"),
		})
	}

	for _, ref := range root.UnknownRefs {
		if ref.Uri != uri {
			continue
		}

		node := ref.Node
		message := L("unknown_person")

		if IsNameRef(node) {
			f := root.FindFamily(ref.Surname)

			if f != nil {
				node = node.NamedChild(1)
			} else {
				node = node.NamedChild(0)
				message = L("unknown_family")
			}
		} else if IsNewSurname(node.Parent()) {
			message = L("unknown_family")
		}

		r, err := doc.NodeToRange(node)

		if err != nil {
			LogDebug("Diagnostic error: %s", err.Error())
			continue
		}

		list = append(list, proto.Diagnostic{
			Severity: P(proto.DiagnosticSeverityError),
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

				r, err := doc.NodeToRange(NameRefName(ref.Node))

				if err != nil {
					LogDebug("Diagnostic error: %s", err.Error())
					continue
				}

				if locations == nil {
					d, err := tempDocs.Get(family.Uri)

					if err != nil {
						LogDebug("Diagnostic error: %s", err.Error())
						continue
					}

					locations = make([]proto.DiagnosticRelatedInformation, count)
					for i, dup := range dups {
						node := dup.Member.Node
						rr, err := d.NodeToRange(node)

						if err != nil {
							LogDebug("Diagnostic error: %s", err.Error())
							continue
						}

						sources := GetClosestSources(node)

						locations[i] = proto.DiagnosticRelatedInformation{
							Location: proto.Location{
								URI:   family.Uri,
								Range: *rr,
							},
							Message: L("child_of_source", ToString(sources, d)),
						}
					}
				}

				list = append(list, proto.Diagnostic{
					Severity:           P(proto.DiagnosticSeverityWarning),
					Range:              *r,
					Message:            L("duplicate_count_of_name", count, name),
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
		Docs:     make(map[string]*Ctx),
		Debounce: debounce.New(200 * time.Millisecond),
	}
}

func (dd *DocDebouncer) Set(uri Uri, ctx *Ctx) {
	dd.Docs[uri] = ctx
	dd.Debounce(func() {
		dd.Flush()
	})
}

func (dd *DocDebouncer) Flush() {
	root.UpdateDirty()

	for uri, ctx := range dd.Docs {
		delete(dd.Docs, uri)

		if !IsFamilyUri(uri) || !UriFileExist(uri) {
			continue
		}

		PublishDiagnostics(ctx, uri, nil)
	}
}
