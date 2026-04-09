package layout

import (
	"sync"

	"github.com/redexp/familymarkup-lsp/state"
	fm "github.com/redexp/familymarkup-parser"
	flex "github.com/redexp/go-flextree"
)

func Align(root *state.Root, params AlignParams) []*SvgFamily {
	graphFamilies := GraphDocumentFamilies(root)

	svgFamilies := make([]*SvgFamily, len(graphFamilies))

	var wg sync.WaitGroup

	// align roots
	for fi, gf := range graphFamilies {
		f := &SvgFamily{
			Uri: gf.uri,
			Loc: gf.Name.Loc(),
			Title: Node{
				Rect: Rect{
					Width:  int(float64(gf.Name.CharsNum) * ss.FamilyTitleSize * params.FontRatio),
					Height: int(ss.FamilyTitleSize),
				},
				Name: gf.Name.Text,
			},
		}

		gf.svgFamily = f
		svgFamilies[fi] = f

		wg.Go(func() {
			tree := &flex.Tree{
				Input:    &GraphPerson{},
				Width:    float64(f.Title.Width) + ss.PersonMarginX*2,
				Height:   float64(ss.LevelHeight),
				Children: make([]*flex.Tree, len(gf.RootPersons)),
			}

			for i, p := range gf.RootPersons {
				tree.Children[i] = createFlexTree(p, params)
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

				if p.graphPerson != nil {
					p.graphPerson.svgPerson = p
				}
			})

			node.Walk(func(p *SvgPerson) {
				p.X += -left + ss.BorderPadding
				//p.Y += ss.BorderPadding
			})

			f.Width = right - left + ss.BorderPadding*2
			f.Height = bottom + ss.BorderPadding*2
			f.Title.X = node.X
			f.Title.Y = ss.LevelHeight - f.Title.Height
			f.Roots = node.Children
		})
	}

	wg.Wait()

	// create levels and bounding
	for _, f := range svgFamilies {
		wg.Go(func() {
			lvlMap := make(LevelsMap)

			lvlMap.Add(Rect{
				X:      f.Title.X,
				Y:      0,
				Width:  f.Title.Width,
				Height: ss.LevelHeight,
			})

			f.Walk(func(p *SvgPerson) {
				rect := p.Rect
				rect.Y -= ss.ArrowsHeight
				rect.Height = ss.LevelHeight

				link := p.graphPerson.Link

				if link != nil {
					f.links = append(f.links, &SvgLink{
						Family: link.Family.svgFamily,
						From:   p.Rect,
						To:     link.svgPerson.Rect,
					})
				}

				lvlMap.Add(rect)
			})

			f.levels = lvlMap.ToArray()
			mergeLevelsRects(f.levels, 160)
			f.Bounding = levelsToBounding(f.levels)
		})
	}

	wg.Wait()

	alignByLevels(svgFamilies)

	for _, family := range svgFamilies {
		family.Walk(func(p *SvgPerson) {
			link := p.graphPerson.Link

			if link == nil {
				return
			}

			svgPerson := link.svgPerson

			p.Pointers = append(p.Pointers, &SvgPointer{
				Label:  p.graphPerson.Person.Surname.Text,
				Family: link.Family.svgFamily.Rect,
				Person: svgPerson.Rect,
			})

			svgPerson.Pointers = append(svgPerson.Pointers, &SvgPointer{
				Label:  family.Title.Name,
				Family: family.Rect,
				Person: p.Rect,
			})
		})
	}

	return svgFamilies
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

		if first != nil {
			tree.Children[i] = first
		} else {
			last = tree
		}

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
		Height: float64(ss.LevelHeight),
	}
}

func flexTreeToSvgPerson(tree *flex.Tree, walk func(*SvgPerson)) *SvgPerson {
	gp := tree.Input.(*GraphPerson)

	p := &SvgPerson{
		Rect: Rect{
			X: int(tree.X + ss.PersonMarginX),
			Y: int(tree.Y) + ss.ArrowsHeight,

			Width:  int(tree.Width - ss.PersonMarginX*2),
			Height: int(ss.PersonHeight),
		},

		graphPerson: gp,
	}

	if token := gp.Token(); token != nil {
		p.Name = token.Text
		p.Loc = token.Loc()
		p.Unknown = token.Type == fm.TokenUnknown
		p.External = !gp.Person.IsChild && gp.Person.Surname != nil
	}

	p.Children = make([]*SvgPerson, len(tree.Children))

	for i, child := range tree.Children {
		p.Children[i] = flexTreeToSvgPerson(child, walk)
	}

	if gp != nil {
		for i, rel := range gp.Relations {
			if len(rel.Partners) == 0 {
				if len(rel.Label) > 0 {
					p.Rel = &SvgRel{
						Label: rel.Label,
					}
				}

				continue
			}

			last := len(rel.Partners) - 1
			svgPartner := p.Children[i]

			for j := range rel.Partners {
				svgRel := &SvgRel{}

				if j < len(rel.Separators) {
					svgRel.Separator = rel.Separators[j]
				}

				if j == last {
					svgRel.Label = rel.Label
				}

				svgPartner.Rel = svgRel

				if len(svgPartner.Children) == 0 {
					break
				}

				svgPartner = svgPartner.Children[0]
			}
		}
	}

	walk(p)

	return p
}
