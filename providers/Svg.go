package providers

import (
	"encoding/json"
	"slices"
	"sync"

	. "github.com/redexp/familymarkup-lsp/types"
	fm "github.com/redexp/familymarkup-parser"
	flex "github.com/redexp/go-flextree"
)

var ss = SvgStyle{
	FamilyTitleSize: 16,
	FamilyPadding:   10,
	FamilyGap:       15,
	PersonNameSize:  12,
	PersonHeight:    30,
	PersonPaddingX:  20,
	PersonMarginX:   10,
	ArrowsHeight:    25,
}

func SvgDocument(_ *Ctx, params *SvgDocumentParams) (list []*SvgFamily, err error) {
	families := GraphDocumentFamilies(NormalizeUri(params.URI))
	list = make([]*SvgFamily, len(families))

	var wg sync.WaitGroup

	// align roots
	for i, family := range families {
		f := &SvgFamily{
			Title: SvgNode{
				Rect: Rect{
					Width:  int(float64(family.Name.CharsNum) * float64(ss.FamilyTitleSize) * params.FontRatio),
					Height: ss.FamilyTitleSize,
				},
				Name: family.Name.Text,
			},
		}

		f.Roots = make([]*SvgRoot, len(family.RootPersons))

		list[i] = f

		for j, p := range family.RootPersons {
			wg.Go(func() {
				tree := createFlexTree(p, params)
				tree.Reset()
				tree.Update()

				left := 0
				right := 0
				bottom := 0

				node := flexTreeToPerson(tree, func(p *SvgPerson) {
					p.X -= p.Width / 2

					left = min(left, p.X)
					right = max(right, p.X+p.Width)
					bottom = max(bottom, p.Y+p.Height)
				})

				node.Walk(func(p *SvgPerson) {
					p.X += -left
				})

				f.Roots[j] = &SvgRoot{
					Rect: Rect{
						Width:  right - left,
						Height: bottom,
					},
					Person: node,
				}
			})
		}
	}

	wg.Wait()

	var walk func(*SvgRoot, *SvgPerson)

	walk = func(r *SvgRoot, p *SvgPerson) {
		p.X += r.X
		p.Y += r.Y

		for _, child := range p.Children {
			walk(r, child)
		}
	}

	// align families
	for _, f := range list {
		wg.Go(func() {
			for ri, r := range f.Roots {
				r.Y = ss.FamilyPadding + ss.FamilyTitleSize + ss.ArrowsHeight

				f.Width += r.Width
				f.Height = max(f.Height, r.Height)

				if ri == 0 {
					r.X += ss.FamilyPadding
					walk(r, r.Person)
					continue
				}

				f.Width += ss.FamilyGap

				prev := f.Roots[ri-1]

				r.X = prev.X + prev.Width + ss.FamilyGap
				walk(r, r.Person)
			}

			f.Width = ss.FamilyPadding + f.Width + ss.FamilyPadding
			f.Height = ss.FamilyPadding + ss.FamilyTitleSize + ss.ArrowsHeight + f.Height + ss.FamilyPadding
			updateSvgFamilyTitle(f)
		})
	}

	wg.Wait()

	// create bounding path
	for _, f := range list {
		count := len(f.Roots)

		if count == 0 {
			continue
		}

		wg.Go(func() {
			var rightPoints []SvgPos

			var g sync.WaitGroup

			g.Go(func() {
				firstRoot := f.Roots[0]
				persons := make(map[int]Rect)

				firstRoot.Person.Walk(func(p *SvgPerson) {
					prev, ok := persons[p.Y]

					if ok && p.X >= prev.X {
						return
					}

					persons[p.Y] = p.Rect
				})

				rects := make([]Rect, 0, len(persons)+1)

				rects = append(rects, f.Title.Rect)

				rects[0].Height += ss.ArrowsHeight

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

				points := make([]SvgPos, 0, len(rects)*2)

				for _, r := range rects {
					points = append(points, r.Pos("tl"), r.Pos("bl"))
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
				lastRoot := f.Roots[count-1]
				persons := make(map[int]Rect)

				lastRoot.Person.Walk(func(p *SvgPerson) {
					prev, ok := persons[p.Y]

					if ok && p.Pos("tr").X <= prev.X {
						return
					}

					persons[p.Y] = p.Move(p.Width, 0)
				})

				rects := make([]Rect, 0, len(persons)+1)

				rects = append(rects, f.Title.Move(f.Title.Width, 0))

				rects[0].Height += ss.ArrowsHeight

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

				points := make([]SvgPos, 0, len(rects)*2)

				for _, p := range rects {
					points = append(points, p.Pos("tl"), p.Pos("bl"))
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

	for i := 1; i < len(list); i++ {
		prev := list[i-1]
		cur := list[i]

		cur.Y = prev.Pos("bl").Y
	}

	return list, nil
}

func createFlexTree(p *GraphPerson, params *SvgDocumentParams) *flex.Tree {
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

func personToFlexTree(p *GraphPerson, params *SvgDocumentParams) *flex.Tree {
	var token *fm.Token

	if p.Person.Unknown != nil {
		token = p.Person.Unknown
	} else {
		token = p.Person.Name
	}

	return &flex.Tree{
		Input:  token,
		Width:  float64(token.CharsNum)*ss.PersonNameSize*params.FontRatio + ss.PersonPaddingX*2 + ss.PersonMarginX*2,
		Height: ss.PersonHeight + float64(ss.ArrowsHeight),
	}
}

func flexTreeToPerson(tree *flex.Tree, walk func(*SvgPerson)) *SvgPerson {
	input := tree.Input.(*fm.Token)

	p := &SvgPerson{
		Rect: Rect{
			X: int(tree.X + ss.PersonMarginX),
			Y: int(tree.Y),

			Width:  int(tree.Width - ss.PersonMarginX*2),
			Height: int(tree.Height) - ss.ArrowsHeight,
		},

		Name: input.Text,
	}

	p.Children = make([]*SvgPerson, len(tree.Children))

	for i, child := range tree.Children {
		p.Children[i] = flexTreeToPerson(child, walk)
	}

	walk(p)

	return p
}

func updateSvgFamilyTitle(f *SvgFamily) {
	count := len(f.Roots)

	if count == 0 {
		return
	}

	first := f.Roots[0]
	last := f.Roots[count-1]

	width := last.Person.X + last.Person.Width - first.Person.X

	f.Title.X = first.Person.X + width/2 - f.Title.Width/2
	f.Title.Y = ss.FamilyPadding
}

type SvgStyle struct {
	FamilyTitleSize int
	FamilyPadding   int
	FamilyGap       int
	PersonNameSize  float64
	PersonHeight    float64
	PersonPaddingX  float64
	PersonMarginX   float64
	ArrowsHeight    int
}

type SvgPos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`

	Width  int `json:"width"`
	Height int `json:"height"`
}

func (r Rect) Pos(t string) SvgPos {
	pos := SvgPos{
		X: r.X,
		Y: r.Y,
	}

	switch t {
	case "tl":
		return pos
	case "tr":
		pos.X += r.Width
	case "bl":
		pos.Y += r.Height
	case "bl+":
		pos.Y += r.Height + ss.ArrowsHeight
	case "br":
		pos.X += r.Width
		pos.Y += r.Height
	case "br+":
		pos.X += r.Width
		pos.Y += r.Height + ss.ArrowsHeight
	default:
		panic("invalid Pos type: " + t)
	}

	return pos
}

func (r Rect) Move(x, y int) Rect {
	r.X += x
	r.Y += y
	return r
}

type SvgNode struct {
	Rect

	Name string `json:"name"`
}

type SvgFamily struct {
	Rect

	Title SvgNode `json:"title"`

	Roots []*SvgRoot `json:"roots"`

	Bounding []SvgPos `json:"bounding"`
}

type SvgRoot struct {
	Rect

	Person *SvgPerson `json:"person"`
}

type SvgPerson struct {
	Rect

	Name string `json:"name"`

	Children []*SvgPerson `json:"children"`
}

func (p *SvgPerson) Walk(cb func(*SvgPerson)) {
	cb(p)

	for _, child := range p.Children {
		child.Walk(cb)
	}
}

type SvgHandlers struct {
	Document SvgDocumentFunc
}

func (req *SvgHandlers) Handle(ctx *Ctx) (res any, validMethod bool, validParams bool, err error) {
	switch ctx.Method {
	case SvgDocumentMethod:
		validMethod = true

		var params SvgDocumentParams
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.Document(ctx, &params)
		}
	}

	return
}

const SvgDocumentMethod = "svg/document"

type SvgDocumentParams struct {
	URI       Uri     `json:"URI"`
	FontRatio float64 `json:"fontRatio"`
}

type SvgDocumentFunc func(*Ctx, *SvgDocumentParams) ([]*SvgFamily, error)
