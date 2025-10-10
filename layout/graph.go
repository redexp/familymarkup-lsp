package layout

import (
	"sync"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	fm "github.com/redexp/familymarkup-parser"
)

func GraphDocumentFamilies(root *Root, uri Uri) []*GraphFamily {
	personMem := make(map[*fm.Person]*Member)

	for _, ref := range root.NodeRefs[uri] {
		switch ref.Type {
		case RefTypeName, RefTypeNameSurname:
			personMem[ref.Person] = ref.Member

		case RefTypeOrigin:
			personMem[ref.Person] = ref.Member.Origin
		}
	}

	memGP := make(map[*Member]*GraphPerson)

	findGP := func(p *fm.Person) (gp *GraphPerson, mem *Member) {
		mem, ok := personMem[p]

		if ok {
			gp = memGP[mem]
		}

		return
	}

	toGP := func(p *fm.Person) *GraphPerson {
		return &GraphPerson{
			Person: p,
		}
	}

	list := make([]*GraphFamily, len(root.Docs[uri].Root.Families))

	//links := make(map[*GraphPerson]*GraphPerson)

	for i, f := range root.Docs[uri].Root.Families {
		gf := &GraphFamily{
			Name: f.Name,
		}

		list[i] = gf

		for _, rel := range f.Relations {
			if !rel.IsFamilyDef {
				continue
			}

			var gp *GraphPerson
			var partners []*GraphPerson
			var mem *Member

			for _, p := range rel.Sources.Persons {
				if gp != nil {
					partner := toGP(p)

					mem = personMem[p]

					if mem != nil {
						memGP[mem] = partner
					}

					partners = append(partners, partner)
					continue
				}

				gp, mem = findGP(p)

				if gp != nil {
					continue
				}

				gp = toGP(p)
				memGP[mem] = gp
				gf.RootPersons = append(gf.RootPersons, gp)
			}

			if gp == nil {
				if len(partners) == 0 {
					continue
				}

				gp = partners[0]
				partners = partners[1:]
			}

			gr := &GraphRelation{
				Partners: partners,
			}

			gp.Relations = append(gp.Relations, gr)

			if rel.Targets == nil {
				continue
			}

			for _, p := range rel.Targets.Persons {
				child := toGP(p)
				gr.Children = append(gr.Children, child)

				mem = personMem[p]

				if mem != nil {
					memGP[mem] = child
				}
			}
		}
	}

	var walk func(*GraphPerson)

	walk = func(p *GraphPerson) {
		for _, rel := range p.Relations {
			for _, partner := range rel.Partners {
				mem := personMem[partner.Person]

				if mem == nil {
					continue
				}

				partner.Link = memGP[mem]
			}

			for _, child := range rel.Children {
				walk(child)
			}
		}
	}

	var wg sync.WaitGroup

	for _, gf := range list {
		for _, gp := range gf.RootPersons {
			wg.Go(func() {
				walk(gp)
			})
		}
	}

	wg.Wait()

	return list
}

type GraphFamily struct {
	Name        *fm.Token
	RootPersons []*GraphPerson
}

type GraphPerson struct {
	Person    *fm.Person
	Link      *GraphPerson
	Relations []*GraphRelation
}

type GraphRelation struct {
	Partners []*GraphPerson
	Children []*GraphPerson
}
