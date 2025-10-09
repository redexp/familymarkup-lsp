package layout

type Pos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`

	Width  int `json:"width"`
	Height int `json:"height"`
}

func (r Rect) ToPos(t string) Pos {
	pos := Pos{
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

	Roots []*SvgRoot `json:"roots"`

	Bounding []Pos `json:"bounding"`
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
