package providers

import (
	"fmt"
	"iter"
	"regexp"
	"strconv"
	"strings"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func DocFormating(ctx *Ctx, params *proto.DocumentFormattingParams) (list []proto.TextEdit, err error) {
	return prettyfy(params.TextDocument.URI, nil)
}

func RangeFormating(ctx *Ctx, params *proto.DocumentRangeFormattingParams) (list []proto.TextEdit, err error) {
	return prettyfy(params.TextDocument.URI, &params.Range)
}

func LineFormating(ctx *Ctx, params *proto.DocumentOnTypeFormattingParams) (list []proto.TextEdit, err error) {
	pos := params.Position
	line := pos.Line

	r := &Range{
		Start: Position{
			Line:      line,
			Character: 0,
		},
		End: Position{
			Line:      line,
			Character: pos.Character,
		},
	}

	newLine := params.Ch == "\n"

	if newLine {
		r.Start.Line--
	}

	list, err = prettyfy(params.TextDocument.URI, r)

	if err != nil {
		return
	}

	if newLine {
		edits, err := addNewLineNum(params.TextDocument.URI, &pos)

		if err != nil {
			return nil, err
		}

		if len(edits) > 0 {
			list = append(list, edits...)
		}
	}

	return
}

func prettyfy(uri Uri, r *Range) (list []proto.TextEdit, err error) {
	doc, err := TempDoc(uri)

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

	validRange := func(a *Range, b *Range) bool {
		if a != nil && a.Start.Line != a.End.Line {
			return false
		}

		if b != nil && b.Start.Line != b.End.Line {
			return false
		}

		if a != nil && b != nil && a.End.Line != b.Start.Line {
			return false
		}

		return true
	}

	checkFirst := func(pos *Range) {
		if pos.Start.Character == 0 {
			return
		}

		add(proto.TextEdit{
			Range: Range{
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
				Range: Range{
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

		if err != nil || !validRange(namePos, aliasesPos) {
			return
		}

		checkBetween(&namePos.End, &aliasesPos.Start, " ")

		prev := aliases.NamedChild(0)

		if prev == nil {
			return
		}

		prevPos, err := doc.NodeToRange(prev)

		if err != nil || !validRange(aliasesPos, prevPos) {
			return
		}

		checkBetween(&aliasesPos.Start, &prevPos.Start, "(")

		for nextPos, err := range childPosIter(prev, doc) {
			if err != nil || !validRange(prevPos, nextPos) {
				return err
			}

			checkBetween(&prevPos.End, &nextPos.Start, ", ")

			prevPos = nextPos
		}

		if validRange(prevPos, aliasesPos) {
			checkBetween(&prevPos.End, &aliasesPos.End, ")")
		}

		return
	}

	for index, node := range QueryIter(q, doc.Tree.RootNode(), []byte(doc.Text)) {
		nodePos, err := doc.NodeToRange(node)

		if err != nil {
			return nil, err
		}

		if r != nil && !RangeOverlaps(r, nodePos) {
			continue
		}

		switch index {
		case 0, 1:
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

				if next.IsMissing() {
					continue
				}

				nextPos, err := doc.NodeToRange(next)

				if err != nil {
					return nil, err
				}

				text := " "

				if !next.IsNamed() && !next.IsError() && ToString(next, doc) == "," {
					text = ""
				}

				if validRange(prevPos, nextPos) {
					checkBetween(&prevPos.End, &nextPos.Start, text)
				}

				prev = next
				prevPos = nextPos
			}

			arrow := node.NextNamedSibling()

			if arrow == nil || arrow.IsMissing() {
				continue
			}

			kind := arrow.Kind()

			if kind != "arrow" && kind != "eq" {
				continue
			}

			arrowPos, err := doc.NodeToRange(arrow)

			if err != nil || !validRange(prevPos, arrowPos) {
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

				if err != nil || !validRange(arrowPos, nodePos) {
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
					Range: Range{
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

func addNewLineNum(uri Uri, pos *Position) (list []proto.TextEdit, err error) {
	doc, err := TempDoc(uri)

	if err != nil {
		return
	}

	text, err := doc.GetTextOnLine(pos.Line - 1)

	if err != nil {
		return
	}

	textLen := uint32(len(text))
	text = strings.TrimSpace(text)

	match, err := regexp.MatchString("^\\d+\\.?$", text)

	if err != nil {
		return
	}

	if match {
		list = append(list, proto.TextEdit{
			Range: Range{
				Start: Position{
					Line:      pos.Line - 1,
					Character: 0,
				},
				End: Position{
					Line:      pos.Line - 1,
					Character: textLen,
				},
			},
			NewText: "",
		})

		return
	}

	exp := regexp.MustCompile(`^(\d+)\.? `)

	replaceNums := func(line uint32, index uint) {
		for {
			text, err := doc.GetTextOnLine(line)

			if err != nil {
				return
			}

			match := exp.FindStringSubmatch(text)

			if len(match) == 0 {
				return
			}

			num, err := strconv.Atoi(match[1])

			if err != nil {
				return
			}

			if num != int(index) {
				list = append(list, proto.TextEdit{
					Range: Range{
						Start: Position{
							Line:      line,
							Character: 0,
						},
						End: Position{
							Line:      line,
							Character: uint32(len(match[0])),
						},
					},
					NewText: fmt.Sprintf("%d. ", index),
				})
			}

			line++
			index++
		}
	}

	match, err = regexp.MatchString(`=[\p{Ll}'" ]*$`, text)

	if err != nil {
		return
	}

	if match {
		list = append(list, proto.TextEdit{
			Range: Range{
				Start: *pos,
				End:   *pos,
			},
			NewText: "1. ",
		})

		replaceNums(pos.Line+1, 2)

		return
	}

	result := exp.FindStringSubmatch(text)

	if len(result) == 0 {
		return
	}

	num, err := strconv.Atoi(result[1])

	if err != nil {
		return
	}

	list = append(list, proto.TextEdit{
		Range: Range{
			Start: *pos,
			End:   *pos,
		},
		NewText: fmt.Sprintf("%d. ", num+1),
	})

	replaceNums(pos.Line+1, uint(num+2))

	return
}

func childPosIter(prev *Node, doc *TextDocument) iter.Seq2[*Range, error] {
	return func(yield func(*Range, error) bool) {
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
