package state

import (
	"math"

	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
)

func addDuplicate(duplicates Duplicates, name string, dup *Duplicate) {
	_, exist := duplicates[name]

	if !exist {
		duplicates[name] = make([]*Duplicate, 0)
	}

	duplicates[name] = append(duplicates[name], dup)
}

func filterRefs(refs []*Ref, uris UriSet) []*Ref {
	list := make([]*Ref, 0)

	for _, ref := range refs {
		if uris.Has(ref.Uri) {
			continue
		}

		list = append(list, ref)
	}

	return list
}

func getAliasesNode(node *Node) *Node {
	next := node.NextNamedSibling()

	if IsNameAliases(next) {
		return next
	}

	parent := node.Parent()

	if IsNameDef(parent) {
		return parent.ChildByFieldName("aliases")
	}

	return nil
}

func getAliases(nameNode *Node, text []byte) []string {
	node := getAliasesNode(nameNode)

	if node == nil {
		return make([]string, 0)
	}

	count := int(node.NamedChildCount())
	list := make([]string, count)

	for i := 0; i < count; i++ {
		list[i] = node.NamedChild(i).Content(text)
	}

	return list
}

func compareNames(a []rune, b []rune) uint {
	al := float64(len(a))
	bl := float64(len(b))
	max := uint(math.Max(al, bl))
	min := uint(math.Min(al, bl))
	diff := uint(max - min)

	if diff > 2 {
		return diff
	}

	for i := uint(0); i < min; i++ {
		if a[i] != b[i] {
			return max - 1 - i
		}
	}

	return diff
}
