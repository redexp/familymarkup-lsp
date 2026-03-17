package layout

import (
	"github.com/redexp/familymarkup-lsp/types"
	fm "github.com/redexp/familymarkup-parser"
)

type Pos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (p Pos) Move(x, y int) Pos {
	p.X += x
	p.Y += y
	return p
}

type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`

	Width  int `json:"width"`
	Height int `json:"height"`
}

func (r Rect) Right() int {
	return r.X + r.Width
}

const (
	TL = "tl"
	TR = "tr"
	BL = "bl"
	BR = "br"
)

func (r Rect) ToPos(t string) Pos {
	pos := Pos{
		X: r.X,
		Y: r.Y,
	}

	switch t {
	case TL:
		return pos
	case "tm":
		pos.X += r.Width / 2
	case TR:
		pos.X += r.Width
	case BL:
		pos.Y += r.Height
	case BR:
		pos.X += r.Width
		pos.Y += r.Height
	case "center":
		pos.X += r.Width / 2
		pos.Y += r.Height / 2
	default:
		panic("invalid ToPos type: " + t)
	}

	return pos
}

func (r Rect) Move(x, y int) Rect {
	r.X += x
	r.Y += y
	return r
}

type Node struct {
	Rect

	Name string `json:"name"`
}

type SvgFamily struct {
	Rect

	Uri      types.Uri    `json:"uri"`
	Loc      fm.Loc       `json:"loc"`
	Title    Node         `json:"title"`
	Roots    []*SvgPerson `json:"roots"`
	Bounding []Pos        `json:"bounding"`
	levels   []*Level
	links    []*SvgLink
}

func (f *SvgFamily) Walk(cb func(*SvgPerson)) {
	for _, person := range f.Roots {
		person.Walk(cb)
	}
}

type SvgPerson struct {
	Rect

	Name string `json:"name"`
	Loc  fm.Loc `json:"loc"`

	graphPerson *GraphPerson
	Link        *Pos `json:"link,omitempty"`

	Children []*SvgPerson `json:"children"`
}

func (p *SvgPerson) Walk(cb func(*SvgPerson)) {
	cb(p)

	for _, child := range p.Children {
		child.Walk(cb)
	}
}

type SvgLink struct {
	Family *SvgFamily
	From   Rect
	To     Rect
}
