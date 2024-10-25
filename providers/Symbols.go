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
	empty := count == 0

	var q string

	if count == 1 {
		q = parts[0]
	} else if count > 1 {
		q = parts[1]
	}

	list = make([]WorkspaceSymbol, 0)

	for f := range root.FamilyIter() {
		fs := WorkspaceSymbol{
			SymbolInformation: proto.SymbolInformation{
				Kind: proto.SymbolKindNamespace,
				Name: f.Name,
			},
			Location: proto.TextDocumentIdentifier{
				URI: f.Uri,
			},
		}

		if empty {
			list = append(list, fs)
		} else {
			valid := false

			for name := range f.NamesIter() {
				if startsWith(name, parts[0]) {
					valid = true
					break
				}
			}

			if !valid && count > 1 {
				continue
			}

			if valid {
				list = append(list, fs)
			}
		}

		for mem := range f.MembersIter() {
			ms := WorkspaceSymbol{
				SymbolInformation: proto.SymbolInformation{
					Kind:          proto.SymbolKindConstant,
					Name:          mem.Name,
					ContainerName: &fs.Name,
				},
				Location: proto.TextDocumentIdentifier{
					URI: f.Uri,
				},
			}

			if empty {
				list = append(list, ms)
				continue
			}

			for name := range mem.NamesIter() {
				if startsWith(name, q) {
					if count > 1 {
						ms.Name = fmt.Sprintf("%s %s", fs.Name, name)
					} else {
						ms.Name = name
					}

					list = append(list, ms)
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

		if len(parts) > 1 {
			name = parts[1]
		}

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
