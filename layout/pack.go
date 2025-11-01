package layout

import (
	"iter"
	"math"
)

type Point struct {
	X, Y int
}

type Figure struct {
	ID       int
	Cells    []Point // Относительные координаты квадратов
	Position Point   // Позиция фигуры в глобальном пространстве
	LinkCell Point   // Относительная позиция точки связи
	LinkTo   *Figure // Связь с другой фигурой
}

func (f *Figure) ToAbs(p Point) Point {
	p.X += f.Position.X
	p.Y += f.Position.Y
	return p
}

func (f *Figure) AbsCellsIter() iter.Seq[Point] {
	return func(yield func(Point) bool) {
		for _, cell := range f.Cells {
			if !yield(f.ToAbs(cell)) {
				return
			}
		}
	}
}

func (f *Figure) AbsCellsMap() (cells map[Point]bool) {
	cells = make(map[Point]bool)

	for p := range f.AbsCellsIter() {
		cells[p] = true
	}

	return
}

func (f *Figure) AbsLinkCell() Point {
	return f.ToAbs(f.LinkCell)
}

func (p Point) Distance(target Point) Point {
	return Point{
		X: target.X - p.X,
		Y: target.Y - p.Y,
	}
}

func (p Point) AbsSum() int {
	return abs(p.X) + abs(p.Y)
}

func (p Point) Iter() iter.Seq2[int, int] {
	return func(yield func(int, int) bool) {
		xd := direction(p.X) * -1
		yd := direction(p.Y) * -1

		for x := range abs(p.X) + 1 {
			for y := range abs(p.Y) + 1 {
				if !yield(p.X+x*xd, p.Y+y*yd) {
					return
				}
			}
		}
	}
}

func Distance(p1, p2 Point) float64 {
	dx := float64(p1.X - p2.X)
	dy := float64(p1.Y - p2.Y)
	return math.Sqrt(dx*dx + dy*dy)
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

func Pack(figures []*Figure) {
	for _, fig := range figures {
		targetFig := fig.LinkTo

		if targetFig == nil {
			continue
		}

		targetCell := targetFig.AbsLinkCell()
		targetCellsMap := targetFig.AbsCellsMap()

		maxDist := fig.AbsLinkCell().Distance(targetCell)
		minDist := maxDist
		origin := fig.Position
		closest := fig.Position

	outer:
		for x, y := range maxDist.Iter() {
			fig.Position.X = origin.X + x
			fig.Position.Y = origin.Y + y

			dist := fig.AbsLinkCell().Distance(targetCell)

			if dist.AbsSum() >= minDist.AbsSum() {
				continue
			}

			for p := range fig.AbsCellsIter() {
				if targetCellsMap[p] {
					continue outer
				}
			}

			minDist = dist
			closest = fig.Position
		}

		fig.Position = closest
	}
}
