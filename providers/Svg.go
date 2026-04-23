package providers

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dominikbraun/graph"
	"github.com/redexp/familymarkup-lsp/layout"
	. "github.com/redexp/familymarkup-lsp/types"
	"github.com/redexp/familymarkup-lsp/utils"
	fm "github.com/redexp/familymarkup-parser"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func SvgFamilies(_ *Ctx, params *SvgFamiliesParams) (SvgFamiliesResult, error) {
	families, relations := layout.Align(root, layout.AlignParams{
		FontRatio: params.FontRatio,
	})

	return SvgFamiliesResult{
		Families:  families,
		Relations: relations,
	}, nil
}

func SvgPath(_ *Ctx, params *SvgPathParams) (res SvgPathResult, err error) {
	if len(params.Persons) < 2 {
		err = fmt.Errorf("should be more than 1 persons")
		return
	}

	type PersonId struct {
		Uri Uri
		Pos fm.Position
	}

	personHash := func(p *layout.GraphPerson) *layout.GraphPerson {
		return p
	}

	g := graph.New(personHash)

	families, _ := layout.CreateGraphFamilies(root)

	persons := make([]*layout.GraphPerson, len(params.Persons))

	found := 0

	edge := func(a *layout.GraphPerson, b *layout.GraphPerson) {
		err := g.AddEdge(a, b)

		if err != nil {
			fmt.Println(err.Error())
		}
	}

	var add func(*layout.GraphPerson)

	add = func(p *layout.GraphPerson) {
		err := g.AddVertex(p)

		if err != nil {
			if !errors.Is(err, graph.ErrVertexAlreadyExists) {
				fmt.Println(err.Error())
			}

			return
		}

		if p.Link != nil {
			add(p.Link)
			edge(p, p.Link)
		}

		if found == len(persons) {
			return
		}

		for i, item := range params.Persons {
			if p.Family.Uri != item.URI {
				continue
			}

			token := p.Token()

			if token.Line == int(item.Position.Line) && token.Char == int(item.Position.Character) {
				persons[i] = p
				found++
				break
			}
		}
	}

	for _, f := range families {
		f.Walk(func(p *layout.GraphPerson) {
			for _, rel := range p.Relations {
				list := make([]*layout.GraphPerson, 0, 1+len(rel.Partners)+len(rel.Children))

				list = append(list, p)
				list = append(list, rel.Partners...)
				list = append(list, rel.Children...)

				for i := 0; i < len(list); i++ {
					add(list[i])

					for j := i + 1; j < len(list); j++ {
						add(list[j])
						edge(list[i], list[j])
					}
				}
			}
		})
	}

	for i, person := range persons {
		if person == nil {
			err = fmt.Errorf("person at index %d - not found", i)
			return
		}
	}

	list, err := graph.ShortestPath(g, persons[0], persons[1])

	if err != nil {
		return
	}

	res.Path = make([]SvgPathPerson, 0, len(list))

	for _, p := range list {
		res.Path = append(res.Path, SvgPathPerson{
			URI:      p.Family.Uri,
			Position: utils.TokenToPosition(p.Person.Name),
		})
	}

	return
}

type SvgHandlers struct {
	Families SvgFamiliesFunc
	Path     SvgPathFunc
}

func (req *SvgHandlers) Handle(ctx *Ctx) (res any, validMethod bool, validParams bool, err error) {
	switch ctx.Method {
	case SvgFamiliesMethod:
		validMethod = true

		var params SvgFamiliesParams
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.Families(ctx, &params)
		}

	case SvgPathMethod:
		validMethod = true

		var params SvgPathParams
		if err = json.Unmarshal(ctx.Params, &params); err == nil {
			validParams = true
			res, err = req.Path(ctx, &params)
		}
	}

	return
}

const SvgFamiliesMethod = "svg/families"

type SvgFamiliesParams struct {
	URI       Uri     `json:"URI"`
	FontRatio float64 `json:"fontRatio"`
}

type SvgFamiliesResult struct {
	Families  []*layout.SvgFamily   `json:"families"`
	Relations []*layout.SvgRelation `json:"relations"`
}

type SvgFamiliesFunc func(*Ctx, *SvgFamiliesParams) (SvgFamiliesResult, error)

const SvgPathMethod = "svg/path"

type SvgPathParams struct {
	Persons []SvgPathPerson `json:"persons"`
}

type SvgPathPerson struct {
	URI      Uri            `json:"uri"`
	Position proto.Position `json:"position"`
}

type SvgPathResult struct {
	Path []SvgPathPerson `json:"path"`
}

type SvgPathFunc func(*Ctx, *SvgPathParams) (SvgPathResult, error)
