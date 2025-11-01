package layout

import (
	"fmt"
	"testing"
)

func TestPack(t *testing.T) {
	fig1 := &Figure{
		ID: 1,
		Cells: []Point{
			{0, 0}, {1, 0}, {2, 0},
			{0, 1},
		},
		Position: Point{0, 0},
		LinkCell: Point{0, 1},
	}

	fig2 := &Figure{
		ID: 2,
		Cells: []Point{
			{0, 0}, {1, 0},
			{0, 1}, {1, 1},
		},
		Position: Point{-10, 5},
		LinkCell: Point{0, 0},
	}

	// Устанавливаем связь
	fig1.LinkTo = fig2

	figures := []*Figure{fig1, fig2}

	fmt.Println("Начальные позиции:")
	printFigures(figures)

	Pack(figures)

	fmt.Println("\nКонечные позиции:")
	printFigures(figures)
}

func printFigures(figures []*Figure) {
	cells := make([][]Point, len(figures))
	leftTop := Point{}
	rightBottom := Point{}

	for i, fig := range figures {
		cells[i] = make([]Point, 0, len(fig.Cells))

		for p := range fig.AbsCellsIter() {
			cells[i] = append(cells[i], p)

			if p.X < leftTop.X {
				leftTop.X = p.X
			}
			if p.Y < leftTop.Y {
				leftTop.Y = p.Y
			}

			if p.X > rightBottom.X {
				rightBottom.X = p.X
			}
			if p.Y > rightBottom.Y {
				rightBottom.Y = p.Y
			}
		}
	}

	has := func(x, y int) bool {
		for _, points := range cells {
			for _, p := range points {
				if p.X == x && p.Y == y {
					return true
				}
			}
		}
		return false
	}

	for y := leftTop.Y; y <= rightBottom.Y; y++ {
		for x := leftTop.X; x <= rightBottom.X; x++ {
			if has(x, y) {
				fmt.Print("O")
			} else {
				fmt.Print("_")
			}
		}
		fmt.Print("\n")
	}

	fmt.Print("\n")

	for _, fig := range figures {
		linkPoint := fig.AbsLinkCell()
		fmt.Printf("Фигура %d: позиция (%d, %d), точка связи (%d, %d)\n",
			fig.ID, fig.Position.X, fig.Position.Y, linkPoint.X, linkPoint.Y)

		if fig.LinkTo != nil {
			linkPoint2 := fig.LinkTo.AbsLinkCell()
			dist := Distance(linkPoint, linkPoint2)
			fmt.Printf("  Расстояние до связанной фигуры: %.2f\n", dist)
		}
	}
}
