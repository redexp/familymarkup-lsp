package src

import (
	"fmt"

	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func Symbol(ctx *glsp.Context, params *proto.DocumentSymbolParams) (res any, err error) {
	uri, err := normalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	waitTreesReady()
	root.UpdateDirty()

	doc, err := tempDoc(uri)

	if err != nil {
		return
	}

	list := make([]proto.DocumentSymbol, 0)

	for f := range root.FamilyIter() {
		if f.Uri != uri {
			continue
		}

		r, err := doc.NodeToRange(getClosestNode(f.Node, "family"))

		if err != nil {
			return nil, err
		}

		sr, err := doc.NodeToRange(f.Node)

		if err != nil {
			return nil, err
		}

		symbol := proto.DocumentSymbol{
			Kind:           proto.SymbolKindNamespace,
			Name:           f.Name,
			Range:          *r,
			SelectionRange: *sr,
			Children:       make([]proto.DocumentSymbol, 0),
		}

		for mem := range f.MembersIter() {
			r, err := doc.NodeToRange(mem.Node)

			if err != nil {
				return nil, err
			}

			symbol.Children = append(symbol.Children, proto.DocumentSymbol{
				Kind:           proto.SymbolKindConstant,
				Name:           mem.Name,
				Detail:         pt(fmt.Sprintf("%s %s", f.Name, mem.Name)),
				Range:          *r,
				SelectionRange: *r,
			})
		}

		list = append(list, symbol)
	}

	return list, nil
}
