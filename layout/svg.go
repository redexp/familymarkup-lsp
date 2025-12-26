package layout

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

func (r Rect) ToPos(t string) Pos {
	pos := Pos{
		X: r.X,
		Y: r.Y,
	}

	switch t {
	case "tl":
		return pos
	case "tm":
		pos.X += r.Width / 2
	case "tr":
		pos.X += r.Width
	case "bl":
		pos.Y += r.Height
	case "br":
		pos.X += r.Width
		pos.Y += r.Height
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

	Title Node `json:"title"`

	Roots []*SvgPerson `json:"roots"`

	Bounding []Pos `json:"bounding"`
}

func (f SvgFamily) Walk(cb func(*SvgPerson)) {
	rootPerson := &SvgPerson{
		Rect:     f.Title.Rect,
		Children: f.Roots,
	}

	rootPerson.Width += int(ss.PersonPaddingX)
	rootPerson.Height += ss.ArrowsHeight

	rootPerson.Walk(cb)
}

type SvgPerson struct {
	Rect

	Name   string `json:"name"`
	person *GraphPerson
	Link   *Pos `json:"link,omitempty"`

	Children []*SvgPerson `json:"children"`
}

func (p *SvgPerson) Walk(cb func(*SvgPerson)) {
	cb(p)

	for _, child := range p.Children {
		child.Walk(cb)
	}
}
