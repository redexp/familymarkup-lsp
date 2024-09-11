package src

import (
	"encoding/json"
	"fmt"
	"iter"
	"slices"
	"strings"
	"time"

	"github.com/bep/debounce"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/tliron/glsp"
	proto "github.com/tliron/glsp/protocol_3_16"
)

var treeContext *glsp.Context
var treeReloadDebouncer = debounce.New(2 * time.Second)

func TreeFamilies(ctx *glsp.Context) ([]*TreeFamily, error) {
	list := make([]*TreeFamily, 0)

	for f := range root.FamilyIter() {
		list = append(list, &TreeFamily{
			TreeItemPoint: TreeItemPoint(f.Node.StartPoint()),

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

func TreeRelations(ctx *glsp.Context, loc *TreeItemLocation) (list []*TreeRelation, err error) {
	f, doc, err := getFamilyDoc(loc)

	if err != nil {
		return
	}

	relIter, err := getRelationsIter(f)

	if err != nil {
		return
	}

	list = make([]*TreeRelation, 0)

	for _, relNode := range relIter {
		sourcesNode := relNode.ChildByFieldName("sources")
		arrowNode := relNode.ChildByFieldName("arrow")

		list = append(list, &TreeRelation{
			TreeItemPoint: TreeItemPoint(sourcesNode.StartPoint()),

			Label: toString(sourcesNode, doc),
			Arrow: toString(arrowNode, doc),
		})
	}

	return
}

func TreeMembers(ctx *glsp.Context, loc *TreeItemLocation) (list []*TreeMember, err error) {
	f, doc, err := getFamilyDoc(loc)

	if err != nil {
		return
	}

	relIter, err := getRelationsIter(f)

	if err != nil {
		return
	}

	var relationNode *Node

	for _, relNode := range relIter {
		if relNode.StartPoint().Row == loc.Row {
			relationNode = relNode
			break
		}
	}

	if relationNode == nil {
		return nil, fmt.Errorf("relation not found")
	}

	targets := relationNode.ChildByFieldName("targets")

	if targets == nil {
		return make([]*TreeMember, 0), nil
	}

	count := int(targets.NamedChildCount())
	list = make([]*TreeMember, count)

	for i := 0; i < count; i++ {
		node := targets.NamedChild(i)

		if isNameDef(node) {
			mem := root.GetMemberByUriNode(f.Uri, node.ChildByFieldName("name"))

			if mem != nil {
				list[i] = &TreeMember{
					TreeItemPoint: TreeItemPoint(mem.Node.StartPoint()),

					Name:    mem.Name,
					Aliases: mem.Aliases,
				}
				continue
			}
		} else if isNumUnknown(node) {
			node = node.NamedChild(1)
		}

		list[i] = &TreeMember{
			TreeItemPoint: TreeItemPoint(node.StartPoint()),

			Name: toString(node, doc),
		}
	}

	return
}

func TreeLocation(ctx *glsp.Context, params *TreeLocationParams) (pos *proto.Position, err error) {
	doc, err := tempDoc(params.URI)

	if err != nil {
		return
	}

	return doc.PointToPosition(sitter.Point{
		Row:    params.Row,
		Column: params.Column,
	})
}

func TreeReload() {
	treeContext.Notify("tree/reload", nil)
}

func getFamilyDoc(loc *TreeItemLocation) (f *Family, doc *TextDocument, err error) {
	doc, err = tempDoc(loc.URI)

	if err != nil {
		return
	}

	dups, exist := root.Duplicates[loc.FamilyName]

	if exist {
		for _, dup := range dups {
			if dup.Family.Node.StartPoint().Row == loc.Row {
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

func getRelationsIter(family *Family) (iter.Seq2[uint32, *Node], error) {
	q, err := createQuery(`
		(relation) @rel
	`)

	if err != nil {
		return nil, err
	}

	return queryIter(q, getClosestNode(family.Node, "family")), nil
}

// TreeHandler

type TreeHandlers struct {
	TreeFamilies  TreeFamiliesFunc
	TreeRelations TreeRelationsFunc
	TreeMembers   TreeMembersFunc
	TreeLocation  TreeLocationFunc
}

func (req *TreeHandlers) Handle(ctx *glsp.Context) (res any, validMethod bool, validParams bool, err error) {
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

	case TreeLocationMethod:
		validMethod = true

		var params TreeLocationParams
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.TreeLocation(ctx, &params)
		}
	}

	return
}

type TreeItemPoint struct {
	Row    uint32 `json:"row"`
	Column uint32 `json:"column"`
}

// TreeFamilies

const TreeFamiliesMethod = "tree/families"

type TreeFamiliesFunc func(ctx *glsp.Context) ([]*TreeFamily, error)

type TreeFamily struct {
	TreeItemPoint

	URI     Uri      `json:"uri"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases,omitempty"`
}

// TreeRelations

const TreeRelationsMethod = "tree/relations"

type TreeRelationsFunc func(ctx *glsp.Context, loc *TreeItemLocation) ([]*TreeRelation, error)

type TreeItemLocation struct {
	URI        Uri    `json:"uri"`
	FamilyName string `json:"family_name"`
	Row        uint32 `json:"row"`
}

type TreeRelation struct {
	TreeItemPoint

	Label string `json:"label"`
	Arrow string `json:"arrow"`
}

// TreeMembers

const TreeMembersMethod = "tree/members"

type TreeMembersFunc func(ctx *glsp.Context, loc *TreeItemLocation) ([]*TreeMember, error)

type TreeMember struct {
	TreeItemPoint

	Name    string   `json:"name"`
	Aliases []string `json:"aliases,omitempty"`
}

// TreeLocation

const TreeLocationMethod = "tree/location"

type TreeLocationFunc func(ctx *glsp.Context, params *TreeLocationParams) (*proto.Position, error)

type TreeLocationParams struct {
	TreeItemPoint

	URI Uri `json:"uri"`
}
