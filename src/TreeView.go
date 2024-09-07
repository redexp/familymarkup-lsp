package src

import (
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/tliron/glsp"
)

func TreeFamilies(ctx *glsp.Context) ([]*TreeFamily, error) {
	list := make([]*TreeFamily, 0)

	for f := range root.FamilyIter() {
		list = append(list, &TreeFamily{
			Id:      f.Id,
			Name:    f.Name,
			Aliases: f.Aliases,
		})
	}

	slices.SortFunc(list, func(a *TreeFamily, b *TreeFamily) int {
		return strings.Compare(a.Name, b.Name)
	})

	return list, nil
}

func TreeRelations(ctx *glsp.Context, params *TreeRelationsParams) (list []*TreeRelation, err error) {
	f, doc, err := getFamilyDoc(params.FamilyId)

	if err != nil {
		return
	}

	relIter, err := getRelationsIter(f)

	if err != nil {
		return
	}

	list = make([]*TreeRelation, 0)
	id := uint32(0)

	for _, relNode := range relIter {
		id += 1
		sourcesNode := relNode.ChildByFieldName("sources")
		arrowNode := relNode.ChildByFieldName("arrow")

		list = append(list, &TreeRelation{
			Id:    id,
			Label: toString(sourcesNode, doc),
			Arrow: toString(arrowNode, doc),
		})
	}

	return
}

func TreeMembers(ctx *glsp.Context, params *TreeMembersParams) (list []*TreeMember, err error) {
	f, doc, err := getFamilyDoc(params.FamilyId)

	if err != nil {
		return
	}

	relIter, err := getRelationsIter(f)

	if err != nil {
		return
	}

	id := uint32(0)
	var relationNode *Node

	for _, relNode := range relIter {
		id += 1

		if id == params.RelationId {
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
					Id:      mem.Id,
					Name:    mem.Name,
					Aliases: mem.Aliases,
				}
				continue
			}
		} else if isNumUnknown(node) {
			node = node.NamedChild(1)
		}

		list[i] = &TreeMember{
			Id:   string(i),
			Name: toString(node, doc),
		}
	}

	return
}

func getFamilyDoc(id string) (*Family, *TextDocument, error) {
	f, exist := root.Families[id]

	if !exist {
		return nil, nil, fmt.Errorf("family not found")
	}

	doc, err := tempDoc(f.Uri)

	if err != nil {
		return nil, nil, err
	}

	return f, doc, nil
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
