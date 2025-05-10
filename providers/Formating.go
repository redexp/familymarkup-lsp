package providers

import (
	"fmt"
	fm "github.com/redexp/familymarkup-parser"
	"regexp"
	"strconv"
	"strings"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	. "github.com/redexp/familymarkup-lsp/utils"
	proto "github.com/tliron/glsp/protocol_3_16"
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
	doc, err := TempDoc(uri)

	if err != nil {
		return
	}

	validRange := func(rng *Range) bool {
		if r == nil {
			return true
		}

		start := rng.Start
		end := rng.End

		if r.Start.Line < start.Line && end.Line < r.End.Line {
			return true
		}

		if r.Start.Line == start.Line && start.Character < r.Start.Character {
			return false
		}

		if r.End.Line == end.Line && r.End.Character <= end.Character {
			return false
		}

		return true
	}

	add := func(item proto.TextEdit) {
		if !validRange(&item.Range) {
			return
		}

		list = append(list, item)
	}

	check := func(edit proto.TextEdit) (err error) {
		text, err := doc.GetTextByRange(&edit.Range)

		if err != nil {
			return
		}

		if text != edit.NewText {
			add(edit)
		}

		return
	}

	checkNameAliases := func(name *fm.Token, aliases []*fm.Token) (err error) {
		if name != nil && name.Type == fm.TokenSurname && name.Char != 0 {
			add(proto.TextEdit{
				Range: Range{
					Start: Position{
						Line:      uint32(name.Line),
						Character: 0,
					},
					End: TokenToPosition(name),
				},
				NewText: "",
			})
		}

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

	for _, family := range doc.Root.Families {
		err = checkNameAliases(family.Name, family.Aliases)

		if err != nil {
			return
		}

		for _, rel := range family.Relations {
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

					switch sep.Type {
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
							Start: TokenToPosition(prev),
							End:   TokenToPosition(rel.Arrow),
						},
						NewText: " ",
					})
				}

				if next != nil && next.Line == rel.Arrow.Line && next.Type != fm.TokenSpace && next.SubType != fm.TokenNewLine {
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
