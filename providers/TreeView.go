package providers

import (
	"encoding/json"
	"fmt"
	fm "github.com/redexp/familymarkup-parser"
	"slices"
	"strings"
	"time"

	"github.com/bep/debounce"
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
)

var treeContext *Ctx
var treeReloadDebouncer = debounce.New(2 * time.Second)

func TreeFamilies(ctx *Ctx) ([]*TreeFamily, error) {
	list := make([]*TreeFamily, 0)

	for f := range root.FamilyIter() {
		list = append(list, &TreeFamily{
			Position: f.Node.Start,

			URI:     f.Uri,
			Name:    f.Name,
			Aliases: f.Aliases,
		})
	}

	slices.SortFunc(list, func(a *TreeFamily, b *TreeFamily) int {
		return strings.Compare(a.Name, b.Name)
	})

	if treeContext == nil {
		root.OnUpdate(func() {
			treeReloadDebouncer(TreeReload)
		})
	}

	treeContext = ctx

	return list, nil
}

func TreeRelations(_ *Ctx, loc *TreeItemLocation) (list []*TreeRelation, err error) {
	f, doc, err := getFamilyDoc(loc)

	if err != nil {
		return
	}

	list = make([]*TreeRelation, 0)

	for _, rel := range f.Node.Relations {
		r := LocToRange(rel.Sources.Loc)

		label := doc.GetTextByRange(r)

		list = append(list, &TreeRelation{
			Position: rel.Loc.Start,

			Label: label,
			Arrow: rel.Arrow.Text,
		})
	}

	return
}

func TreeMembers(_ *Ctx, loc *TreeItemLocation) (list []*TreeMember, err error) {
	f, _, err := getFamilyDoc(loc)

	if err != nil {
		return
	}

	var relationNode *fm.Relation
	row := int(loc.Row)

	for _, rel := range f.Node.Relations {
		if rel.Start.Line == row {
			relationNode = rel
			break
		}
	}

	if relationNode == nil {
		return nil, fmt.Errorf("relation not found")
	}

	targets := relationNode.Targets

	if targets == nil {
		return make([]*TreeMember, 0), nil
	}

	list = make([]*TreeMember, 0)

	add := func(person *fm.Person, name string, aliases []string) {
		list = append(list, &TreeMember{
			Position: person.Start,

			Name:    name,
			Aliases: aliases,
		})
	}

	for _, person := range targets.Persons {
		if person.Unknown != nil {
			add(person, person.Unknown.Text, []string{})
			continue
		}

		mem := root.GetMemberByUriToken(f.Uri, person.Name)

		if mem != nil {
			add(person, mem.Name, mem.Aliases)
			continue
		}

		add(person, person.Name.Text, TokensToStrings(person.Aliases))
	}

	return
}

func TreeReload() {
	treeContext.Notify("tree/reload", nil)
}

func getFamilyDoc(loc *TreeItemLocation) (f *Family, doc *Doc, err error) {
	doc, err = TempDoc(loc.URI)

	if err != nil {
		return
	}

	dups, exist := root.Duplicates[loc.FamilyName]

	if exist {
		row := int(loc.Row)

		for _, dup := range dups {
			if dup.Family.Node.Start.Line == row {
				f = dup.Family
				return
			}
		}
	}

	f, exist = root.Families[loc.FamilyName]

	if !exist {
		return nil, nil, fmt.Errorf("family not found")
	}

	return
}

// TreeHandler

type TreeHandlers struct {
	TreeFamilies  TreeFamiliesFunc
	TreeRelations TreeRelationsFunc
	TreeMembers   TreeMembersFunc
}

func (req *TreeHandlers) Handle(ctx *Ctx) (res any, validMethod bool, validParams bool, err error) {
	switch ctx.Method {
	case TreeFamiliesMethod:
		validMethod = true
		validParams = true
		res, err = req.TreeFamilies(ctx)

	case TreeRelationsMethod:
		validMethod = true

		var params TreeItemLocation
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.TreeRelations(ctx, &params)
		}

	case TreeMembersMethod:
		validMethod = true

		var params TreeItemLocation
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.TreeMembers(ctx, &params)
		}

	}

	return
}

type TreeItemPoint struct {
	Line uint32 `json:"line"`
	Char uint32 `json:"char"`
}

// TreeFamilies

const TreeFamiliesMethod = "tree/families"

type TreeFamiliesFunc func(ctx *Ctx) ([]*TreeFamily, error)

type TreeFamily struct {
	fm.Position

	URI     Uri      `json:"uri"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases,omitempty"`
}

// TreeRelations

const TreeRelationsMethod = "tree/relations"

type TreeRelationsFunc func(ctx *Ctx, loc *TreeItemLocation) ([]*TreeRelation, error)

type TreeItemLocation struct {
	URI        Uri    `json:"uri"`
	FamilyName string `json:"family_name"`
	Row        uint32 `json:"row"`
}

type TreeRelation struct {
	fm.Position

	Label string `json:"label"`
	Arrow string `json:"arrow"`
}

// TreeMembers

const TreeMembersMethod = "tree/members"

type TreeMembersFunc func(ctx *Ctx, loc *TreeItemLocation) ([]*TreeMember, error)

type TreeMember struct {
	fm.Position

	Name    string   `json:"name"`
	Aliases []string `json:"aliases,omitempty"`
}
