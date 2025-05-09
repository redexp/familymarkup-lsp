package providers

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocSymbols(ctx *Ctx, params *proto.DocumentSymbolParams) (res any, err error) {
	uri, err := NormalizeUri(params.TextDocument.URI)

	if err != nil {
		return
	}

	root.UpdateDirty()

	doc, err := TempDoc(uri)

	if err != nil {
		return
	}

	list := make([]proto.DocumentSymbol, 0)

	for f := range root.FamilyIter() {
		if f.Uri != uri {
			continue
		}

		r, err := doc.NodeToRange(GetClosestNode(f.Node, "family"))

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
				Detail:         P(fmt.Sprintf("%s %s", f.Name, mem.Name)),
				Range:          *r,
				SelectionRange: *r,
			})
		}

		list = append(list, symbol)
	}

	return list, nil
}

func AllSymbols(ctx *Ctx, params *proto.WorkspaceSymbolParams) (list []WorkspaceSymbol, err error) {
	parts := splitQuery(params.Query)
	count := len(parts)

	list = make([]WorkspaceSymbol, 0)

	add := func(uri Uri, name string, container *string) {
		list = append(list, WorkspaceSymbol{
			SymbolInformation: proto.SymbolInformation{
				Kind:          proto.SymbolKindConstant,
				Name:          name,
				ContainerName: container,
			},
			Location: proto.TextDocumentIdentifier{
				URI: uri,
			},
		})
	}

	if count == 0 {
		for f := range root.FamilyIter() {
			add(f.Uri, f.Name, nil)

			for mem := range f.MembersIter() {
				add(f.Uri, mem.Name, &f.Name)
			}
		}

		return
	}

	surnameQuery := parts[0]

	if count > 1 {
		surnameQuery = parts[1]
	}

	for f := range root.FamilyIter() {
		surname := ""

		for name := range f.NamesIter() {
			if startsWith(name, surnameQuery) {
				surname = name
				break
			}
		}

		if count == 1 && surname != "" {
			add(f.Uri, surname, nil)
			continue
		}

		for mem := range f.MembersIter() {
			for name := range mem.NamesIter() {
				if startsWith(name, parts[0]) {
					title := name

					if surname != "" {
						title = fmt.Sprintf("%s %s", name, surname)
					}

					add(f.Uri, title, &f.Name)
					break
				}
			}
		}
	}

	return
}

func ResolveSymbol(ctx *Ctx, symbol *WorkspaceSymbol) (res *WorkspaceSymbolLocation, err error) {
	res = &WorkspaceSymbolLocation{
		SymbolInformation: proto.SymbolInformation{
			Kind:          symbol.Kind,
			Name:          symbol.Name,
			ContainerName: symbol.ContainerName,
		},
	}

	getFamily := func(name string) (*Family, *TextDocument, error) {
		f, exist := root.Families[name]

		if !exist {
			return nil, nil, fmt.Errorf("family not found")
		}

		doc, err := TempDoc(f.Uri)

		if err != nil {
			return nil, nil, err
		}

		return f, doc, nil
	}

	switch symbol.Kind {
	case proto.SymbolKindNamespace:
		f, doc, err := getFamily(symbol.Name)

		if err != nil {
			return nil, err
		}

		r, err := doc.NodeToRange(f.Node)

		if err != nil {
			return nil, err
		}

		res.Location = proto.Location{
			URI:   f.Uri,
			Range: *r,
		}

	case proto.SymbolKindConstant:
		f, doc, err := getFamily(*symbol.ContainerName)

		if err != nil {
			return nil, err
		}

		parts := strings.Split(symbol.Name, " ")
		name := parts[0]

		mem, exist := f.Members[name]

		if !exist {
			return nil, fmt.Errorf("member not found")
		}

		r, err := doc.NodeToRange(mem.Node)

		if err != nil {
			return nil, err
		}

		res.Location = proto.Location{
			URI:   f.Uri,
			Range: *r,
		}
	}

	return
}

func splitQuery(query string) []string {
	query = strings.Trim(query, " ")

	list := make([]string, 0)

	if query == "" {
		return list
	}

	chars := []rune(query)
	from := 0

	for i, r := range chars {
		if r == ' ' || unicode.IsUpper(r) {
			if i == 0 {
				continue
			}

			str := strings.Trim(string(chars[from:i]), " ")

			if str != "" {
				list = append(list, strings.ToLower(str))
			}

			from = i
		}
	}

	list = append(list, strings.ToLower(string(chars[from:])))

	return list
}

func startsWith(name string, q string) bool {
	return strings.HasPrefix(strings.ToLower(name), q)
}

type WorkspaceSymbolOptions struct {
	ResolveProvider bool `json:"resolveProvider"`
}

type WorkspaceSymbol struct {
	proto.SymbolInformation

	Location proto.TextDocumentIdentifier `json:"location"`
	Data     any                          `json:"data,omitempty"`
}

type WorkspaceSymbolLocation struct {
	proto.SymbolInformation

	Location proto.Location `json:"location"`
}

type WorkspaceHandler struct {
	WorkspaceSymbol        WorkspaceSymbolFunc
	WorkspaceSymbolResolve WorkspaceSymbolResolveFunc
}

func (req *WorkspaceHandler) Handle(ctx *Ctx) (res any, validMethod bool, validParams bool, err error) {
	switch ctx.Method {
	case proto.MethodWorkspaceSymbol:
		validMethod = true

		var params proto.WorkspaceSymbolParams
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.WorkspaceSymbol(ctx, &params)
		}

	case MethodWorkspaceSymbolResolve:
		validMethod = true

		var params WorkspaceSymbol
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.WorkspaceSymbolResolve(ctx, &params)
		}
	}

	return
}

type WorkspaceSymbolFunc func(ctx *Ctx, params *proto.WorkspaceSymbolParams) ([]WorkspaceSymbol, error)

const MethodWorkspaceSymbolResolve = "workspaceSymbol/resolve"

type WorkspaceSymbolResolveFunc func(ctx *Ctx, symbol *WorkspaceSymbol) (*WorkspaceSymbolLocation, error)
