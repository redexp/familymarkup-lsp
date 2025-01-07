package providers

import (
	"fmt"
	"iter"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocFormating(ctx *Ctx, params *proto.DocumentFormattingParams) (list []proto.TextEdit, err error) {
	doc, err := TempDoc(params.TextDocument.URI)

	if err != nil {
		return
	}

	q, err := CreateQuery(`
		(family_name) @f
		(name_def) @n

		(sources) @s
		
		(targets) @t
	`)

	if err != nil {
		return
	}

	defer q.Close()

	add := func(items ...proto.TextEdit) {
		list = append(list, items...)
	}

	checkFirst := func(pos *proto.Range) {
		if pos.Start.Character == 0 {
			return
		}

		add(proto.TextEdit{
			Range: proto.Range{
				Start: Position{
					Line:      pos.Start.Line,
					Character: 0,
				},
				End: pos.Start,
			},
			NewText: "",
		})
	}

	checkBetween := func(aPos *Position, bPos *Position, text string) {
		if bPos.Character-aPos.Character != uint32(len(text)) {
			add(proto.TextEdit{
				Range: proto.Range{
					Start: *aPos,
					End:   *bPos,
				},
				NewText: text,
			})
		}
	}

	checkNameAliases := func(node *Node) (err error) {
		name := node.ChildByFieldName("name")
		namePos, err := doc.NodeToRange(name)

		if err != nil {
			return
		}

		if IsFamilyName(node) {
			checkFirst(namePos)
		}

		aliases := node.ChildByFieldName("aliases")

		if aliases == nil {
			return
		}

		aliasesPos, err := doc.NodeToRange(aliases)

		if err != nil {
			return
		}

		checkBetween(&namePos.End, &aliasesPos.Start, " ")

		prev := aliases.NamedChild(0)

		if prev == nil {
			return
		}

		prevPos, err := doc.NodeToRange(prev)

		if err != nil {
			return
		}

		checkBetween(&aliasesPos.Start, &prevPos.Start, "(")

		for nextPos, err := range childPosIter(prev, doc) {
			if err != nil {
				return err
			}

			checkBetween(&prevPos.End, &nextPos.Start, ", ")

			prevPos = nextPos
		}

		checkBetween(&prevPos.End, &aliasesPos.End, ")")

		return
	}

	for index, node := range QueryIter(q, doc.Tree.RootNode(), []byte(doc.Text)) {
		switch index {
		case 0:
			checkNameAliases(node)
		case 1:
			checkNameAliases(node)

		case 2:
			prev := node.NamedChild(0)
			prevPos, err := doc.NodeToRange(prev)

			if err != nil {
				return nil, err
			}

			checkFirst(prevPos)

			for {
				next := prev.NextSibling()

				if next == nil {
					break
				}

				nextPos, err := doc.NodeToRange(next)

				if err != nil {
					return nil, err
				}

				text := " "

				if !next.IsNamed() && ToString(next, doc) == "," {
					text = ""
				}

				checkBetween(&prevPos.End, &nextPos.Start, text)

				prev = next
				prevPos = nextPos
			}

			arrow := node.NextNamedSibling()
			kind := arrow.Kind()

			if arrow == nil || (kind != "arrow" && kind != "eq") {
				continue
			}

			arrowPos, err := doc.NodeToRange(arrow)

			if err != nil {
				return nil, err
			}

			checkBetween(&prevPos.End, &arrowPos.Start, " ")

		case 3:
			arrow := node.PrevNamedSibling()

			if arrow != nil && arrow.StartPosition().Row == node.StartPosition().Row {
				arrowPos, err := doc.NodeToRange(arrow)

				if err != nil {
					return nil, err
				}

				nodePos, err := doc.NodeToRange(node)

				if err != nil {
					return nil, err
				}

				checkBetween(&arrowPos.End, &nodePos.Start, " ")
			}

			hasNum := false

			for child := range ChildrenIter(node) {
				hasNum = child.ChildByFieldName("number") != nil

				if hasNum {
					break
				}
			}

			if !hasNum {
				continue
			}

			num := 0
			prevLine := node.PrevNamedSibling().EndPosition().Row

			for child := range ChildrenIter(node) {
				kind := child.Kind()
				childLine := child.StartPosition().Row
				sameLine := prevLine == childLine
				prevLine = childLine

				if kind != "name_def" && kind != "num_unknown" && kind != "unknown" {
					continue
				}

				num++

				numNode := child.ChildByFieldName("number")
				numText := fmt.Sprintf("%d.", num)

				if numNode != nil && ToString(numNode, doc) == numText {
					continue
				}

				nameNode := child.ChildByFieldName("name")

				if nameNode == nil {
					nameNode = child
				}

				namePos, err := doc.NodeToRange(nameNode)

				if err != nil {
					return nil, err
				}

				start := uint32(0)

				if sameLine {
					start = namePos.Start.Character
				}

				add(proto.TextEdit{
					Range: proto.Range{
						Start: Position{
							Line:      namePos.Start.Line,
							Character: start,
						},
						End: namePos.Start,
					},
					NewText: numText + " ",
				})
			}
		}
	}

	return
}

func childPosIter(prev *Node, doc *TextDocument) iter.Seq2[*proto.Range, error] {
	return func(yield func(*proto.Range, error) bool) {
		for {
			next := prev.NextNamedSibling()

			if next == nil {
				break
			}

			nextPos, err := doc.NodeToRange(next)

			if !yield(nextPos, err) {
				break
			}

			prev = next
		}
	}
}
