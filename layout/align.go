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

		wg.Go(func() {
			leftRects := make(map[int]Rect)
			rightRects := make(map[int]Rect)

			f.Walk(func(p *SvgPerson) {
				rect := p.Rect
				level := rect.Y

				prev, ok := leftRects[level]

				if !ok || rect.X < prev.X {
					leftRects[level] = rect
				}

				prev, ok = rightRects[level]

				if !ok || rect.ToPos("tr").X > prev.X {
					rightRects[level] = rect.Move(p.Width, 0)
				}
			})

			var rightPoints []Pos

			var g sync.WaitGroup

			g.Go(func() {
				rects := make([]Rect, 0, len(leftRects))

				for _, rect := range leftRects {
					rects = append(rects, rect)
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

				f.Bounding = points
			})

			g.Go(func() {
				rects := make([]Rect, 0, len(rightRects))

				for _, rect := range rightRects {
					rects = append(rects, rect)
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

				slices.Reverse(points)

				rightPoints = points
			})

			g.Wait()

			f.Bounding = append(f.Bounding, rightPoints...)

			addBoundingPadding(f)
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

	if token == nil {
		panic("token nil")
	}

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

func addBoundingPadding(f *SvgFamily) {
	points := f.Bounding
	last := len(points) - 1
	list := make([]Pos, last+1)
	p := ss.FamilyPadding

	list[0] = points[0].Move(-p, -p)
	list[last] = points[last].Move(p, -p)

	for i := 1; i < last; i++ {
		prev := points[i-1]
		cur := points[i]
		next := points[i+1]
		x := 0
		y := 0

		if prev == cur {
			list[i] = list[i-1]
			continue
		}

		if prev.X == cur.X && prev.Y < cur.Y {
			x = -1
			if next.X < cur.X {
				y = -1
			} else if next.X > cur.X {
				y = 1
			}
		} else if prev.Y == cur.Y && cur.X < prev.X {
			y = -1
			if next.Y < cur.Y {
				x = 1
			} else if next.Y > cur.Y {
				x = -1
			}
		} else if prev.Y == cur.Y && prev.X < cur.X {
			y = 1
			if next.Y < cur.Y {
				x = 1
			} else if next.Y > cur.Y {
				x = -1
			}
		} else if prev.X == cur.X && prev.Y > cur.Y {
			x = 1
			if next.X < cur.X {
				y = -1
			} else if next.X > cur.X {
				y = 1
			}
		}

		list[i] = cur.Move(p*x, p*y)
	}

	f.Bounding = list
}

//func boundingBetweenPoints(leftPoints, rightPoints []Pos, bottomRects []Rect) (list []Pos) {
//	if len(leftPoints) == 0 || len(rightPoints) == 0 {
//		return
//	}
//
//	left := leftPoints[len(leftPoints)-1]
//	right := rightPoints[0]
//
//	if right.X-left.X < 100 {
//		return
//	}
//
//	var rects []Rect
//	prev := &left
//
//	for _, rect := range bottomRects {
//		br := rect.ToPos("br")
//
//		if br.X < prev.X || br == *prev {
//			continue
//		}
//
//		bl := rect.ToPos("bl")
//
//		if bl.X < prev.X {
//			bl.X = prev.X
//		}
//
//	}
//}
