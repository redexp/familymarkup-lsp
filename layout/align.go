package layout

import (
	"slices"
	"sync"

	"github.com/redexp/familymarkup-lsp/state"
	"github.com/redexp/familymarkup-lsp/types"
	flex "github.com/redexp/go-flextree"
)

func Align(root *state.Root, uri types.Uri, params AlignParams) []*SvgFamily {
	families := GraphDocumentFamilies(root, uri)
	list := make([]*SvgFamily, len(families))
	graphFamilies := make(map[*GraphFamily]*SvgFamily)

	var wg sync.WaitGroup

	// align roots
	for i, gf := range families {
		f := &SvgFamily{
			Title: Node{
				Rect: Rect{
					Width:  int(float64(gf.Name.CharsNum) * ss.FamilyTitleSize * params.FontRatio),
					Height: int(ss.FamilyTitleSize),
				},
				Name: gf.Name.Text,
			},
		}

		graphFamilies[gf] = f

		list[i] = f

		wg.Go(func() {
			tree := &flex.Tree{
				Input:    &GraphPerson{},
				Width:    float64(f.Title.Width) + ss.PersonPaddingX*2,
				Height:   float64(f.Title.Height + ss.ArrowsHeight),
				Children: make([]*flex.Tree, 0, len(gf.RootPersons)),
			}

			for _, p := range gf.RootPersons {
				tree.Children = append(tree.Children, createFlexTree(p, params))
			}

			tree.Reset()
			tree.Update()

			left := 0
			right := 0
			bottom := 0

			node := flexTreeToSvgPerson(tree, func(p *SvgPerson) {
				p.X -= p.Width / 2

				left = min(left, p.X)
				right = max(right, p.X+p.Width)
				bottom = max(bottom, p.Y+p.Height)
			})

			node.Walk(func(p *SvgPerson) {
				p.X += -left + ss.FamilyPadding
				p.Y += ss.FamilyPadding
			})

			f.Width = right - left + ss.FamilyPadding*2
			f.Height = bottom + ss.FamilyPadding*2
			f.Title.X = node.X
			f.Title.Y = node.Y
			f.Roots = node.Children
		})
	}

	wg.Wait()

	// create bounding path
	for _, f := range list {
		if len(f.Roots) == 0 {
			continue
		}

		rootPerson := &SvgPerson{
			Rect:     f.Title.Rect,
			Children: f.Roots,
		}

		rootPerson.Width += int(ss.PersonPaddingX)
		rootPerson.Height += ss.ArrowsHeight

		wg.Go(func() {
			var rightPoints []Pos

			var g sync.WaitGroup

			g.Go(func() {
				persons := make(map[int]Rect)

				rootPerson.Walk(func(p *SvgPerson) {
					prev, ok := persons[p.Y]

					if ok && p.X >= prev.X {
						return
					}

					persons[p.Y] = p.Rect
				})

				rects := make([]Rect, 0, len(persons))

				for _, p := range persons {
					rects = append(rects, p)
				}

				slices.SortFunc(rects, func(a, b Rect) int {
					return a.Y - b.Y
				})

				for i, r := range rects {
					if i == 0 {
						continue
					}

					prev := rects[i-1]
					delta := prev.X - r.X

					if -10 < delta && delta < 0 {
						rects[i].X = prev.X
					} else if 0 < delta && delta < 10 {
						rects[i-1].X = r.X
					}

					if i <= 1 {
						continue
					}

					rects[i].Y -= ss.ArrowsHeight
					rects[i].Height += ss.ArrowsHeight
				}

				points := make([]Pos, 0, len(rects)*2)

				for _, r := range rects {
					points = append(points, r.ToPos("tl"), r.ToPos("bl"))
				}

				pad := ss.FamilyPadding

				prev := points[0]

				count := len(points)

				for i := 1; i < count-1; i++ {
					cur := points[i]

					if prev.X == cur.X {
						next := points[i+1]

						if cur.X < next.X {
							cur.Y += pad
						}
					} else if prev.X < cur.X {
						cur.Y += pad
					}

					cur.X -= pad

					prev = points[i]
					points[i] = cur
				}

				points[0].X -= pad
				points[0].Y -= pad
				points[count-1].X -= pad
				points[count-1].Y += pad

				if points[2].X < points[1].X {
					points[1].Y -= pad
					points[2].Y -= pad
				}

				f.Bounding = points
			})

			g.Go(func() {
				persons := make(map[int]Rect)

				rootPerson.Walk(func(p *SvgPerson) {
					prev, ok := persons[p.Y]

					if ok && p.ToPos("tr").X <= prev.X {
						return
					}

					persons[p.Y] = p.Move(p.Width, 0)
				})

				rects := make([]Rect, 0, len(persons))

				for _, p := range persons {
					rects = append(rects, p)
				}

				slices.SortFunc(rects, func(a, b Rect) int {
					return a.Y - b.Y
				})

				for i, r := range rects {
					if i == 0 {
						continue
					}

					prev := rects[i-1]
					delta := prev.X - r.X

					if -10 < delta && delta < 0 {
						rects[i-1].X = r.X
					} else if 0 < delta && delta < 10 {
						rects[i].X = prev.X
					}

					if i <= 1 {
						continue
					}

					rects[i].Y -= ss.ArrowsHeight
					rects[i].Height += ss.ArrowsHeight
				}

				points := make([]Pos, 0, len(rects)*2)

				for _, p := range rects {
					points = append(points, p.ToPos("tl"), p.ToPos("bl"))
				}

				pad := ss.FamilyPadding

				prev := points[0]

				count := len(points)

				for i := 1; i < count-1; i++ {
					cur := points[i]

					if prev.X == cur.X {
						next := points[i+1]

						if next.X < cur.X {
							cur.Y += pad
						}
					} else if cur.X < prev.X {
						cur.Y += pad
					}

					cur.X += pad

					prev = points[i]
					points[i] = cur
				}

				points[0].X += pad
				points[0].Y -= pad
				points[count-1].X += pad
				points[count-1].Y += pad

				if points[1].X < points[2].X {
					points[1].Y -= pad
					points[2].Y -= pad
				}

				slices.Reverse(points)

				rightPoints = points
			})

			g.Wait()

			f.Bounding = append(f.Bounding, rightPoints...)
		})
	}

	wg.Wait()

	alignFamilies(list, graphFamilies)

	return list
}

type AlignParams struct {
	FontRatio float64
}

func createFlexTree(p *GraphPerson, params AlignParams) *flex.Tree {
	tree := personToFlexTree(p, params)

	tree.Children = make([]*flex.Tree, len(p.Relations))

	for i, rel := range p.Relations {
		var first *flex.Tree
		var last *flex.Tree

		for _, partner := range rel.Partners {
			node := personToFlexTree(partner, params)

			if first == nil {
				first = node
			} else if last != nil {
				last.Children = append(last.Children, node)
			}

			last = node
		}

		if first == nil {
			first = &flex.Tree{
				Width:  ss.PersonHeight,
				Height: ss.PersonHeight,
			}

			last = first
		}

		tree.Children[i] = first

		last.Children = make([]*flex.Tree, len(rel.Children))

		for j, child := range rel.Children {
			last.Children[j] = createFlexTree(child, params)
		}
	}

	return tree
}

func personToFlexTree(p *GraphPerson, params AlignParams) *flex.Tree {
	token := p.Token()

	return &flex.Tree{
		Input:  p,
		Width:  float64(token.CharsNum)*ss.PersonNameSize*params.FontRatio + ss.PersonPaddingX*2 + ss.PersonMarginX*2,
		Height: ss.PersonHeight + float64(ss.ArrowsHeight),
	}
}

func flexTreeToSvgPerson(tree *flex.Tree, walk func(*SvgPerson)) *SvgPerson {
	gp := tree.Input.(*GraphPerson)

	p := &SvgPerson{
		Rect: Rect{
			X: int(tree.X + ss.PersonMarginX),
			Y: int(tree.Y),

			Width:  int(tree.Width - ss.PersonMarginX*2),
			Height: int(tree.Height) - ss.ArrowsHeight,
		},

		person: gp,
	}

	if token := gp.Token(); token != nil {
		p.Name = token.Text
	}

	p.Children = make([]*SvgPerson, len(tree.Children))

	for i, child := range tree.Children {
		p.Children[i] = flexTreeToSvgPerson(child, walk)
	}

	walk(p)

	return p
}
