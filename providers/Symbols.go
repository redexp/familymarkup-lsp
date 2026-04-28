package providers

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocSymbols(_ *Ctx, params *proto.DocumentSymbolParams) (res any, err error) {
	uri := NormalizeUri(params.TextDocument.URI)

	err = root.UpdateDirty()

	if err != nil {
		return
	}

	list := make([]proto.DocumentSymbol, 0)

	for f := range root.FamiliesByUriIter(uri) {
		symbol := proto.DocumentSymbol{
			Kind:           proto.SymbolKindNamespace,
			Name:           f.Name,
			Range:          LocToRange(f.Node.Loc),
			SelectionRange: LocToRange(f.Node.Name.Loc()),
			Children:       make([]proto.DocumentSymbol, 0),
		}

		for mem := range f.MembersIter() {
			r := LocToRange(mem.Person.Name.Loc())

			symbol.Children = append(symbol.Children, proto.DocumentSymbol{
				Kind:           proto.SymbolKindConstant,
				Name:           mem.Name,
				Detail:         P(fmt.Sprintf("%s %s", f.Name, mem.Name)),
				Range:          r,
				SelectionRange: r,
			})
		}

		list = append(list, symbol)
	}

	return list, nil
}

func AllSymbols(_ *Ctx, params *WorkspaceSymbolParams) (list []SymbolInformation, err error) {
	memberKind := proto.SymbolKindField

	defer func() {
		if err != nil {
			return
		}

		if list == nil {
			list = make([]SymbolInformation, 0)
			return
		}

		if len(list) < 2 {
			return
		}

		slices.SortFunc(list, func(a, b SymbolInformation) int {
			dir := strings.Compare(a.Name, b.Name)

			if dir == 0 && a.Kind == memberKind && b.Kind == memberKind {
				return strings.Compare(*a.ContainerName, *b.ContainerName)
			}

			return dir
		})

		indexes := make(map[int]struct{})

		for i := 1; i < len(list); i++ {
			item := list[i]
			prev := list[i-1]

			if item.Kind == memberKind &&
				prev.Kind == memberKind &&
				item.Name == prev.Name &&
				*item.ContainerName == *prev.ContainerName {
				indexes[i] = struct{}{}
				indexes[i-1] = struct{}{}
			}
		}

		for index := range indexes {
			p := list[index].member.Person

			if p.Side == fm.SideTargets {
				list[index].Details = L("child_of_source", p.Relation.Sources.Format())
			} else {
				list[index].Details = p.Relation.Sources.Format()
			}
		}
	}()

	parts := splitQuery(params.Query)
	count := len(parts)

	addFamily := func(f *Family, name string) {
		if params.OnlyMembers {
			return
		}

		list = append(list, SymbolInformation{
			SymbolInformation: proto.SymbolInformation{
				Kind: proto.SymbolKindConstant,
				Name: name,
				Location: proto.Location{
					URI:   f.Uri,
					Range: LocToRange(f.Node.Name.Loc()),
				},
			},
		})
	}

	addMember := func(f *Family, mem *Member, name string, surname string) {
		list = append(list, SymbolInformation{
			member: mem,
			SymbolInformation: proto.SymbolInformation{
				Kind:          memberKind,
				Name:          name,
				ContainerName: &surname,
				Location: proto.Location{
					URI:   f.Uri,
					Range: LocToRange(mem.Person.Name.Loc()),
				},
			},
		})
	}

	if count == 0 {
		for f := range root.FamilyIter() {
			addFamily(f, f.Name)

			for mem := range f.MembersIter() {
				addMember(f, mem, mem.Name, f.Name)
			}
		}

		return
	}

	surnameQuery := parts[count-1]

	for f := range root.FamilyIter() {
		surname := ""

		for name := range f.NamesIter() {
			if startsWith(name, surnameQuery) {
				surname = name
				break
			}
		}

		if count == 1 && surname != "" {
			addFamily(f, surname)
			continue
		}

		if params.ExactMatch && count > 1 && surname == "" {
			continue
		}

		for mem := range f.MembersIter() {
			for name := range mem.NamesIter() {
				if !startsWith(name, parts[0]) {
					continue
				}

				if params.ExactMatch && surname != "" {
					addMember(f, mem, name, surname)
					break
				}

				title := name

				if surname != "" {
					title = fmt.Sprintf("%s %s", name, surname)
				}

				addMember(f, mem, title, f.Name)

				break
			}
		}
	}

	return
}

func ResolveSymbol(_ *Ctx, symbol *WorkspaceSymbol) (res *proto.SymbolInformation, err error) {
	res = &proto.SymbolInformation{
		Kind:          symbol.Kind,
		Name:          symbol.Name,
		ContainerName: symbol.ContainerName,
	}

	getFamily := func(name string) (*Family, error) {
		f, exist := root.Families[name]

		if !exist {
			return nil, fmt.Errorf("family not found")
		}

		return f, nil
	}

	var f *Family

	if symbol.Kind == proto.SymbolKindNamespace || (symbol.Kind == proto.SymbolKindConstant && symbol.ContainerName == nil) {
		f, err = getFamily(symbol.Name)

		if err != nil {
			return
		}

		res.Location = proto.Location{
			URI:   f.Uri,
			Range: LocToRange(f.Node.Loc),
		}
	} else if symbol.Kind == proto.SymbolKindConstant {
		f, err = getFamily(*symbol.ContainerName)

		if err != nil {
			return
		}

		parts := strings.Split(symbol.Name, " ")
		name := parts[0]

		mem, exist := f.Members[name]

		if !exist {
			return nil, fmt.Errorf("member not found")
		}

		res.Location = proto.Location{
			URI:   f.Uri,
			Range: LocToRange(mem.Person.Name.Loc()),
		}
	}

	return
}

func MemberSymbol(_ *Ctx, params *MemberSymbolParams) (res *proto.SymbolInformation, err error) {
	uri := NormalizeUri(params.URI)

	ref := root.GetRefByPosition(uri, params.Position)

	if ref == nil || ref.Member == nil {
		return
	}

	mem := ref.Member

	if mem.Origin != nil {
		mem = mem.Origin
	}

	res = &proto.SymbolInformation{
		Kind:          proto.SymbolKindConstant,
		Name:          mem.Name,
		ContainerName: P(mem.Family.Name),
		Location: proto.Location{
			URI:   mem.Family.Uri,
			Range: LocToRange(mem.Person.Name.Loc()),
		},
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

	Data any `json:"data,omitempty"`
}

type SymbolInformation struct {
	proto.SymbolInformation

	Details string `json:"details,omitempty"`
	member  *Member
}

type WorkspaceHandler struct {
	WorkspaceSymbol        WorkspaceSymbolFunc
	WorkspaceSymbolResolve WorkspaceSymbolResolveFunc
	MemberSymbol           MemberSymbolFunc
}

func (req *WorkspaceHandler) Handle(ctx *Ctx) (res any, validMethod bool, validParams bool, err error) {
	switch ctx.Method {
	case proto.MethodWorkspaceSymbol:
		validMethod = true

		var params WorkspaceSymbolParams
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

	case MethodMemberSymbol:
		validMethod = true

		var params MemberSymbolParams
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.MemberSymbol(ctx, &params)
		}
	}

	return
}

type WorkspaceSymbolFunc func(ctx *Ctx, params *WorkspaceSymbolParams) ([]SymbolInformation, error)

type WorkspaceSymbolParams struct {
	proto.WorkspaceSymbolParams

	ExactMatch  bool `json:"exactMatch"`
	OnlyMembers bool `json:"onlyMembers"`
}

const MethodWorkspaceSymbolResolve = "workspaceSymbol/resolve"

type WorkspaceSymbolResolveFunc func(ctx *Ctx, symbol *WorkspaceSymbol) (*proto.SymbolInformation, error)

const MethodMemberSymbol = "workspaceSymbol/member"

type MemberSymbolFunc func(ctx *Ctx, params *MemberSymbolParams) (*proto.SymbolInformation, error)

type MemberSymbolParams struct {
	URI      Uri            `json:"URI"`
	Position proto.Position `json:"position"`
}
