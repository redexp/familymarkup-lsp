package layout

import "slices"

func alignFamilies(list []*SvgFamily, graphFamilies map[*GraphFamily]*SvgFamily) {
	graphPersons := make(map[*GraphPerson]*SvgPerson)

	for _, f := range list {
		for _, p := range f.Roots {
			p.Walk(func(p *SvgPerson) {
				graphPersons[p.person] = p
			})
		}
	}

	famLinks := make(map[*SvgFamily]map[*SvgFamily][]*SvgPerson)

	for _, f := range list {
		for _, p := range f.Roots {
			p.Walk(func(p *SvgPerson) {
				link := p.person.Link

				if link == nil {
					return
				}

				targetFam := graphFamilies[link.Family]

				if targetFam == f {
					return
				}

				if _, ok := famLinks[f]; !ok {
					famLinks[f] = make(map[*SvgFamily][]*SvgPerson)
				}

				if _, ok := famLinks[f][targetFam]; ok {
					return
				}

				if _, ok := famLinks[targetFam][f]; ok {
					return
				}

				targetPers := graphPersons[link]

				famLinks[f][targetFam] = []*SvgPerson{p, targetPers}
			})
		}
	}

	families := make([]*AlignFamily, 0, len(famLinks))

	for f, links := range famLinks {
		fam := &AlignFamily{
			Index:  slices.Index(list, f),
			Family: f,
			Links:  make([]*AlignLink, 0, len(links)),
		}

		families = append(families, fam)

		for targetFam, persons := range links {
			fam.Links = append(fam.Links, &AlignLink{
				From:     persons[0],
				ToFamily: targetFam,
				ToPerson: persons[1],
			})
		}

		slices.SortFunc(fam.Links, func(aLink, bLink *AlignLink) int {
			a := aLink.From.Rect
			b := bLink.From.Rect

			if a.Y != b.Y {
				return a.Y - b.Y
			}

			return a.X - b.X
		})
	}

	slices.SortFunc(families, func(a, b *AlignFamily) int {
		return a.Index - b.Index
	})

	for _, af := range families {
		target := af.Family.ToFigure()

		for _, link := range af.Links {
			target.SetLinkCell(link.From.ToPos("tl"))
			fig := link.ToFamily.ToFigure()
			fig.SetLinkCell(link.ToPerson.ToPos("tl"))
			fig.MoveTo(target)

			link.ToFamily.X = fig.X * ss.GridStep
			link.ToFamily.Y = fig.Y * ss.GridStep

			for _, cell := range fig.Cells {
				target.Cells = append(target.Cells, Pos{
					X: cell.X + fig.X - target.X,
					Y: cell.Y + fig.Y - target.Y,
				})
			}
		}
	}
}

type AlignFamily struct {
	Index  int
	Family *SvgFamily
	Links  []*AlignLink
}

type AlignLink struct {
	From     *SvgPerson
	ToFamily *SvgFamily
	ToPerson *SvgPerson
}
