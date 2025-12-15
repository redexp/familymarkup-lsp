package layout

import (
	"iter"
	"math"
	"sync"
)

type Figure struct {
	Pos

	Cells    []Pos
	LinkCell Pos
}

func (f *Figure) MoveTo(target *Figure) {
	cell := target.AbsLinkCell()
	cellsMap := target.AbsCellsMap()

	type Res struct {
		Pos
		dist  int
		found bool
	}

	check := func(x, y, minDist int) (res Res) {
		link := Pos{X: x, Y: y}

		res.found = false
		res.dist = link.Distance(cell).AbsSum()

		if res.dist >= minDist {
			return
		}

		res.X = x - f.LinkCell.X
		res.Y = y - f.LinkCell.Y

		for p := range CellsIter(f.Cells, res.Pos) {
			if cellsMap[p] {
				return
			}
		}

		res.found = true

		return
	}

	tr := cell
	tl := cell
	br := cell
	bl := cell

	closest := make([]Res, 4)

	for {
		tl = tl.Move(-1, -1)
		tr = tr.Move(1, -1)
		br = br.Move(1, 1)
		bl = bl.Move(-1, 1)

		var wg sync.WaitGroup

		wg.Go(func() {
			prevDist := math.MaxInt

			for y := range rangeIter(tr.Y+1, br.Y-1) {
				res := check(tr.X, y, prevDist)

				if !res.found {
					continue
				}

				prevDist = res.dist
				closest[0] = res
			}
		})

		wg.Go(func() {
			prevDist := math.MaxInt

			for x := range rangeIter(br.X, bl.X) {
				res := check(x, br.Y, prevDist)

				if !res.found {
					continue
				}

				prevDist = res.dist
				closest[1] = res
			}
		})

		wg.Go(func() {
			prevDist := math.MaxInt

			for y := range rangeIter(tl.Y+1, bl.Y-1) {
				res := check(tl.X, y, prevDist)

				if !res.found {
					continue
				}

				prevDist = res.dist
				closest[2] = res
			}
		})

		wg.Go(func() {
			prevDist := math.MaxInt

			for x := range rangeIter(tl.X, tr.X) {
				res := check(x, tl.Y, prevDist)

				if !res.found {
					continue
				}

				prevDist = res.dist
				closest[3] = res
			}
		})

		wg.Wait()

		minDist := math.MaxInt
		index := -1

		for i, res := range closest {
			if res.found && res.dist < minDist {
				minDist = res.dist
				index = i
			}
		}

		if index == -1 {
			continue
		}

		f.Pos = closest[index].Pos

		break
	}
}

func (f *Figure) ToAbs(p Pos) Pos {
	p.X += f.X
	p.Y += f.Y
	return p
}

func (f *Figure) AbsCellsIter() iter.Seq[Pos] {
	return CellsIter(f.Cells, f.Pos)
}

func CellsIter(cells []Pos, pos Pos) iter.Seq[Pos] {
	return func(yield func(Pos) bool) {
		for _, cell := range cells {
			cell.X += pos.X
			cell.Y += pos.Y
			if !yield(cell) {
				return
			}
		}
	}
}

func (f *Figure) AbsCellsMap() (cells map[Pos]bool) {
	cells = make(map[Pos]bool)

	for p := range CellsIter(f.Cells, f.Pos) {
		cells[p] = true
	}

	return
}

func (f *Figure) AbsLinkCell() Pos {
	return f.ToAbs(f.LinkCell)
}

func (f *Figure) SetLinkCell(p Pos) {
	f.LinkCell = Pos{
		X: p.X / ss.GridStep,
		Y: p.Y / ss.GridStep,
	}
}

func (f *Figure) Rect() Rect {
	width := 0
	height := 0

	for _, cell := range f.Cells {
		if cell.X > width {
			width = cell.X
		}

		if cell.Y > height {
			height = cell.Y
		}
	}

	return Rect{
		Width:  width + 1,
		Height: height + 1,
	}
}

func (p Pos) Distance(target Pos) Pos {
	return Pos{
		X: target.X - p.X,
		Y: target.Y - p.Y,
	}
}

func (p Pos) AbsSum() int {
	return abs(p.X) + abs(p.Y)
}

func (p Pos) ToZeroIter() iter.Seq2[int, int] {
	return func(yield func(int, int) bool) {
		for y := range rangeIter(p.Y, 0) {
			for x := range rangeIter(p.X, 0) {
				if !yield(x, y) {
					return
				}
			}
		}
	}
}

func (f *SvgFamily) ToFigure() *Figure {
	step := ss.GridStep

	fig := &Figure{
		Pos: Pos{
			X: f.X / step,
			Y: f.Y / step,
		},
	}

	hash := make(map[Pos]bool)
	var prev *Pos

	for _, p := range f.Bounding {
		cell := Pos{
			X: p.X / step,
			Y: p.Y / step,
		}

		if hash[cell] {
			continue
		}

		hash[cell] = true

		if prev != nil {
			for x := range rangeIter(prev.X, cell.X) {
				if x == prev.X || x == cell.X {
					continue
				}

				fig.Cells = append(fig.Cells, Pos{
					X: x,
					Y: prev.Y,
				})
			}

			for y := range rangeIter(prev.Y, cell.Y) {
				if y == prev.Y || y == cell.Y {
					continue
				}

				fig.Cells = append(fig.Cells, Pos{
					X: prev.X,
					Y: y,
				})
			}
		}

		fig.Cells = append(fig.Cells, cell)

		prev = &cell
	}

	for _, root := range f.Roots {
		root.Walk(func(p *SvgPerson) {
			fromX := p.X / step
			toX := p.ToPos("tr").X / step
			y := p.Y / step

			for x := range rangeIter(fromX, toX) {
				cell := Pos{
					X: x,
					Y: y,
				}

				if hash[cell] {
					return
				}

				hash[cell] = true

				fig.Cells = append(fig.Cells, cell)
			}
		})
	}

	return fig
}

func direction(x int) int {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	}
	return 0
}

func abs(i int) int {
	if i < 0 {
		i = -i
	}

	return i
}

func rangeIter(from, to int) iter.Seq2[int, int] {
	return func(yield func(int, int) bool) {
		d := direction(to - from)
		i := 0

		for n := from; n != to; n += d {
			if !yield(n, i) {
				return
			}
			i++
		}

		yield(to, i)
	}
}
