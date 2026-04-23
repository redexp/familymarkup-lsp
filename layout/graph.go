package layout

import (
	"sync"

	. "github.com/redexp/familymarkup-lsp/state"
	"github.com/redexp/familymarkup-lsp/types"
	fm "github.com/redexp/familymarkup-parser"
)

func CreateGraphFamilies(root *Root) ([]*GraphFamily, []*GraphRelation) {
	personMem := make(map[*fm.Person]*Member)

	for ref := range root.RefsIter() {
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

	var list []*GraphFamily

	for f, doc := range root.FmFamilyIter() {
		gf := &GraphFamily{
			Uri:  doc.Uri,
			Name: f.Name,
		}

		list = append(list, gf)

		for _, rel := range f.Relations {
			if !rel.IsFamilyDef {
				continue
			}

			var mainPerson *fm.Person
			restPersons := make([]*fm.Person, 0, len(rel.Sources.Persons))

			for _, p := range rel.Sources.Persons {
				if !isValidPerson(p) {
					continue
				}

				if mainPerson == nil && p.Surname == nil {
					mainPerson = p
					continue
				}

				restPersons = append(restPersons, p)
			}

			if mainPerson == nil {
				if len(restPersons) == 0 {
					continue
				}

				mainPerson = restPersons[0]
				restPersons = restPersons[1:]
			}

			var gp *GraphPerson
			partners := make([]*GraphPerson, len(restPersons))
			var mem *Member

			gp, mem = findGP(mainPerson)

			if gp == nil {
				gp = createGraphPerson(gf, mainPerson)
				memGP[mem] = gp
				gf.RootPersons = append(gf.RootPersons, gp)
			}

			for i, p := range restPersons {
				partner := createGraphPerson(gf, p)

				mem = personMem[p]

				if mem != nil {
					if _, ok := memGP[mem]; !ok {
						memGP[mem] = partner
					}
				}

				partners[i] = partner
			}

			gr := &GraphRelation{
				Partners: partners,
			}

			for _, sep := range rel.Sources.Separators {
				gr.Separators = append(gr.Separators, sep.Text)
			}

			if rel.Label != nil {
				gr.Label = rel.Label.Text
			}

			gp.Relations = append(gp.Relations, gr)

			if rel.Targets == nil {
				continue
			}

			for _, p := range rel.Targets.Persons {
				if !isValidPerson(p) {
					continue
				}

				child := createGraphPerson(gf, p)
				gr.Children = append(gr.Children, child)

				mem = personMem[p]

				if mem != nil {
					memGP[mem] = child
				}
			}
		}
	}

	setLink := func(p *GraphPerson) {
		mem := personMem[p.Person]

		if mem == nil {
			return
		}

		if link, ok := memGP[mem]; ok && p != link {
			p.Link = link
		}
	}

	var walk func(*GraphFamily, *GraphPerson)

	walk = func(gf *GraphFamily, p *GraphPerson) {
		setLink(p)

		for _, rel := range p.Relations {
			for _, partner := range rel.Partners {
				setLink(partner)
			}

			for _, child := range rel.Children {
				walk(gf, child)
			}
		}
	}

	var wg sync.WaitGroup

	for _, gf := range list {
		for _, gp := range gf.RootPersons {
			wg.Go(func() {
				walk(gf, gp)
			})
		}
	}

	wg.Wait()

	var relations []*GraphRelation

	for _, doc := range root.Docs {
		for _, f := range doc.Root.Families {
			for _, rel := range f.Relations {
				if rel.IsFamilyDef {
					continue
				}

				gr := &GraphRelation{
					Partners: make([]*GraphPerson, 0, len(rel.Sources.Persons)),
				}

				if rel.Label != nil {
					gr.Label = rel.Label.Text
				}

				for _, p := range rel.Sources.Persons {
					gp, _ := findGP(p)

					if gp == nil {
						continue
					}

					gr.Partners = append(gr.Partners, gp)
				}

				if rel.Targets != nil {
					gr.Children = make([]*GraphPerson, 0, len(rel.Targets.Persons))

					for _, p := range rel.Targets.Persons {
						gp, _ := findGP(p)

						if gp == nil {
							continue
						}

						gr.Children = append(gr.Children, gp)
					}
				}

				if len(gr.Partners)+len(gr.Children) < 2 {
					continue
				}

				relations = append(relations, gr)
			}
		}
	}

	return list, relations
}

type GraphFamily struct {
	Name        *fm.Token
	RootPersons []*GraphPerson

	Uri       types.Uri
	svgFamily *SvgFamily
}

func (f *GraphFamily) Walk(cb func(*GraphPerson)) {
	for _, p := range f.RootPersons {
		p.Walk(cb)
	}
}

type GraphPerson struct {
	Family    *GraphFamily
	Person    *fm.Person
	Link      *GraphPerson
	Relations []*GraphRelation

	svgPerson *SvgPerson
}

func (p *GraphPerson) Token() (token *fm.Token) {
	if p.Person == nil {
		return nil
	}

	if p.Person.Unknown != nil {
		token = p.Person.Unknown
	} else {
		token = p.Person.Name
	}

	return
}

func (p *GraphPerson) Walk(cb func(*GraphPerson)) {
	cb(p)

	for _, rel := range p.Relations {
		for _, child := range rel.Children {
			child.Walk(cb)
		}
	}
}

type GraphRelation struct {
	Separators []string
	Label      string
	Partners   []*GraphPerson
	Children   []*GraphPerson
}

func createGraphPerson(f *GraphFamily, p *fm.Person) *GraphPerson {
	return &GraphPerson{
		Family: f,
		Person: p,
	}
}

func isValidPerson(p *fm.Person) bool {
	return p.Name != nil || p.Unknown != nil
}
