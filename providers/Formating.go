package providers

import (
	"fmt"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"
	proto "github.com/tliron/glsp/protocol_3_16"
	"strconv"
	"strings"
)

func DocFormating(_ *Ctx, params *proto.DocumentFormattingParams) (list []proto.TextEdit, err error) {
	return prettify(params.TextDocument.URI, nil)
}

func RangeFormating(_ *Ctx, params *proto.DocumentRangeFormattingParams) (list []proto.TextEdit, err error) {
	return prettify(params.TextDocument.URI, &params.Range)
}

func LineFormating(_ *Ctx, params *proto.DocumentOnTypeFormattingParams) (list []proto.TextEdit, err error) {
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

	list, err = prettify(params.TextDocument.URI, r)

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

func prettify(uri Uri, r *Range) (list []proto.TextEdit, err error) {
	doc, err := GetDoc(uri)

	if err != nil {
		return
	}

	loc := doc.Root.Loc

	if r != nil {
		loc = RangeToLoc(*r)
	}

	add := func(item proto.TextEdit) {
		if !loc.Overlaps(RangeToLoc(item.Range)) {
			return
		}

		list = append(list, item)
	}

	check := func(edit proto.TextEdit) (err error) {
		text := doc.GetTextByRange(edit.Range)

		if text != edit.NewText {
			add(edit)
		}

		return
	}

	checkNameAliases := func(name *fm.Token, aliases []*fm.Token) (err error) {
		if aliases == nil {
			return
		}

		count := len(aliases)

		if count == 0 {
			return
		}

		first := aliases[0]

		if name != nil {
			err = check(proto.TextEdit{
				Range: Range{
					Start: TokenEndToPosition(name),
					End:   TokenToPosition(first),
				},
				NewText: " (",
			})

			if err != nil {
				return
			}
		}

		prev := first

		for i := 1; i < count; i++ {
			alias := aliases[i]

			err = check(proto.TextEdit{
				Range: Range{
					Start: TokenEndToPosition(prev),
					End:   TokenToPosition(alias),
				},
				NewText: ", ",
			})

			if err != nil {
				return
			}

			prev = alias
		}

		last := aliases[count-1]
		lastIndex := doc.TokenIndex(last)
		tokensCount := len(doc.Tokens)

		for i := lastIndex + 1; i < tokensCount; i++ {
			token := doc.Tokens[i]

			if token.Type == fm.TokenSpace {
				continue
			}

			if token.SubType != fm.TokenBracketRight || i-lastIndex == 1 {
				break
			}

			add(proto.TextEdit{
				Range: Range{
					Start: TokenEndToPosition(last),
					End:   TokenToPosition(token),
				},
				NewText: ")",
			})

			break
		}

		return
	}

	for i := loc.Start.Line; i <= loc.End.Line; i++ {
		tokens, ok := doc.TokensByLine[i]

		if !ok {
			continue
		}

		count := len(tokens)

		if count <= 1 {
			continue
		}

		first := tokens[0]
		next := tokens[1]

		if first.Type != fm.TokenSpace {
			continue
		}

		if next.Type == fm.TokenName || next.Type == fm.TokenNum {
			add(proto.TextEdit{
				Range:   TokenToRange(first),
				NewText: "",
			})
		}
	}

	for _, family := range doc.Root.Families {
		switch family.Loc.OverlapType(loc) {
		case fm.OverlapBefore:
			continue
		case fm.OverlapAfter:
			return
		}

		err = checkNameAliases(family.Name, family.Aliases)

		if err != nil {
			return
		}

		for _, rel := range family.Relations {
			switch rel.Loc.OverlapType(loc) {
			case fm.OverlapBefore:
				continue
			case fm.OverlapAfter:
				return
			}

			for _, relList := range []*fm.RelList{rel.Sources, rel.Targets} {
				if relList == nil {
					continue
				}

				hasNum := false

				for _, person := range relList.Persons {
					hasNum = person.Num != nil
					if hasNum {
						break
					}
				}

				for n, person := range relList.Persons {
					err = checkNameAliases(person.Name, person.Aliases)

					if err != nil {
						return
					}

					if person.Num != nil {
						prev, next := doc.PrevNextTokens(person.Num)

						if prev != nil && prev.Type == fm.TokenSpace && prev.Char == 0 {
							add(proto.TextEdit{
								Range:   TokenToRange(prev),
								NewText: "",
							})
						}

						num := strconv.Itoa(n+1) + "."

						if person.Num.Text != num {
							add(proto.TextEdit{
								Range:   TokenToRange(person.Num),
								NewText: num,
							})
						}

						if next != nil && next.Type != fm.TokenSpace {
							add(proto.TextEdit{
								Range: Range{
									Start: TokenEndToPosition(person.Num),
									End:   TokenToPosition(next),
								},
								NewText: " ",
							})
						}
					}
				}

				for _, sep := range relList.Separators {
					prev, next := doc.PrevNextTokens(sep)

					switch sep.SubType {
					case fm.TokenComma:
						if prev != nil && prev.Type == fm.TokenSpace {
							add(proto.TextEdit{
								Range: Range{
									Start: TokenToPosition(prev),
									End:   TokenToPosition(sep),
								},
								NewText: "",
							})
						}

						if next != nil && next.Type != fm.TokenSpace {
							add(proto.TextEdit{
								Range: Range{
									Start: TokenEndToPosition(sep),
									End:   TokenToPosition(next),
								},
								NewText: " ",
							})
						}

					case fm.TokenPlus:
						if prev != nil && prev.Type != fm.TokenSpace {
							add(proto.TextEdit{
								Range: Range{
									Start: TokenEndToPosition(prev),
									End:   TokenToPosition(sep),
								},
								NewText: " ",
							})
						}

						if next != nil && next.Type != fm.TokenSpace {
							add(proto.TextEdit{
								Range: Range{
									Start: TokenEndToPosition(sep),
									End:   TokenToPosition(next),
								},
								NewText: " ",
							})
						}
					}
				}
			}

			if rel.Arrow != nil {
				prev, next := doc.PrevNextTokens(rel.Arrow)

				if prev != nil && prev.Type != fm.TokenSpace {
					add(proto.TextEdit{
						Range: Range{
							Start: TokenEndToPosition(prev),
							End:   TokenToPosition(rel.Arrow),
						},
						NewText: " ",
					})
				}

				if next != nil && next.Line == rel.Arrow.Line && next.Type != fm.TokenSpace && next.SubType != fm.TokenNL {
					add(proto.TextEdit{
						Range: Range{
							Start: TokenEndToPosition(rel.Arrow),
							End:   TokenToPosition(next),
						},
						NewText: " ",
					})
				}
			}
		}
	}

	return
}

func addNewLineNum(uri Uri, pos *Position) (list []proto.TextEdit, err error) {
	doc, err := GetDoc(uri)

	if err != nil {
		return
	}

	tokens := doc.GetTrimTokensByLine(int(pos.Line - 1))
	count := len(tokens)
	var first *fm.Token
	var last *fm.Token

	if count > 0 {
		first = tokens[0]
		last = tokens[count-1]
	}

	if first != nil && first.Type == fm.TokenNum && count == 1 {
		list = append(list, proto.TextEdit{
			Range: Range{
				Start: Position{
					Line:      pos.Line - 1,
					Character: 0,
				},
				End: Position{
					Line:      pos.Line,
					Character: 0,
				},
			},
			NewText: "\n",
		})

		return
	}

	replaceNums := func(line uint32, index int) {
		p := doc.FindPersonByLine(int(line))

		if p == nil || p.Side != fm.SideTargets {
			return
		}

		for _, item := range p.Relation.Targets.Persons {
			if item.Index < p.Index {
				continue
			}

			token := item.Num

			if token == nil {
				index++
				continue
			}

			num, err := numToInt(token)

			if err != nil {
				return
			}

			if num != index {
				list = append(list, proto.TextEdit{
					Range: Range{
						Start: Position{
							Line:      uint32(token.Line),
							Character: 0,
						},
						End: Position{
							Line:      uint32(token.Line),
							Character: uint32(token.EndChar()),
						},
					},
					NewText: fmt.Sprintf("%d. ", index),
				})
			}

			index++
		}
	}

	if last != nil && (last.SubType == fm.TokenEqual || last.Type == fm.TokenWord) {
		hasEq := last.SubType == fm.TokenEqual

		if last.Type == fm.TokenWord {
			prev, _ := doc.PrevNextNonSpaceTokens(last)

			hasEq = prev != nil && prev.SubType == fm.TokenEqual
		}

		if hasEq {
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
	}

	if first == nil || first.Type != fm.TokenNum {
		return
	}

	num, err := numToInt(first)

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

	replaceNums(pos.Line+1, num+2)

	return
}

func numToInt(num *fm.Token) (int, error) {
	return strconv.Atoi(strings.TrimSuffix(num.Text, "."))
}
