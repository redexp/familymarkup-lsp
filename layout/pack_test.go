package layout

import (
	"fmt"
	"testing"
)

func TestToFigure(t *testing.T) {
	root := testRoot(t)

	list := Align(root, "file:///home/sergii/projects/relatives/Ключник/Ключник.family", AlignParams{
		FontRatio: 1,
	})

	f := list[1]

	fig := f.ToFigure()

	if fig == nil {
		t.Fatal("fig == nil")
	}

}

func TestPack(t *testing.T) {
	fig1 := &Figure{
		Cells: []Pos{
			{0, 0}, {1, 0}, {2, 0},
			{0, 1},
		},
		LinkCell: Pos{0, 1},
	}

	fig2 := &Figure{
		Pos: Pos{-10, 5},
		Cells: []Pos{
			{0, 0}, {1, 0},
			{0, 1}, {1, 1},
		},
		LinkCell: Pos{0, 0},
	}

	figures := []*Figure{fig1, fig2}

	printFigures(figures)

	fig1.MoveTo(fig2)

	fmt.Println("Result")
	printFigures(figures)

	if fig1.AbsLinkCell().Distance(fig2.AbsLinkCell()).AbsSum() != 1 {
		t.Error("distance != 1")
	}
}

func printFigures(figures []*Figure) {
	cells := make([][]Pos, len(figures))
	leftTop := Pos{}
	rightBottom := Pos{}

	for i, fig := range figures {
		cells[i] = make([]Pos, 0, len(fig.Cells))

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
}
