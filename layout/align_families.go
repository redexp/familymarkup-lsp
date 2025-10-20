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

	for _, af := range families {
		f := af.Family
		c := f.Width / 2

		for _, link := range af.Links {
			fromPers := link.From
			toPers := link.ToPerson
			toFam := link.ToFamily

			if fromPers.ToPos("tm").X < c {
				toFam.X = f.X - toFam.Width
			} else {
				toFam.X = f.X + f.Width
			}

			toFam.Y = f.Y + (fromPers.Y - toPers.Y)
		}
	}
}

type AlignFamily struct {
	Family *SvgFamily
	Links  []*AlignLink
}

type AlignLink struct {
	From     *SvgPerson
	ToFamily *SvgFamily
	ToPerson *SvgPerson
}
