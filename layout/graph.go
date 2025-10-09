package layout

import (
	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
	fm "github.com/redexp/familymarkup-parser"
)

func GraphDocumentFamilies(root *Root, uri Uri) (list []*GraphFamily) {
	personMem := make(map[*fm.Person]*Member)

	for _, ref := range root.NodeRefs[uri] {
		if ref.Type != RefTypeName {
			continue
		}

		personMem[ref.Person] = ref.Member
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

	for _, f := range root.Docs[uri].Root.Families {
		gf := &GraphFamily{
			Name: f.Name,
		}

		for _, rel := range f.Relations {
			if !rel.IsFamilyDef {
				continue
			}

			var gp *GraphPerson
			var partners []*GraphPerson

			for _, p := range rel.Sources.Persons {
				if gp == nil {
					var mem *Member

					gp, mem = findGP(p)

					if gp != nil {
						continue
					}

					gp = toGP(p)
					memGP[mem] = gp
					gf.RootPersons = append(gf.RootPersons, gp)
					continue
				}

				partners = append(partners, toGP(p))
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

				mem, ok := personMem[p]

				if ok {
					memGP[mem] = child
				}
			}
		}

		list = append(list, gf)
	}

	return
}

type GraphFamily struct {
	Name        *fm.Token
	RootPersons []*GraphPerson
}

type GraphPerson struct {
	Person    *fm.Person
	Relations []*GraphRelation
}

type GraphRelation struct {
	Partners []*GraphPerson
	Children []*GraphPerson
}
