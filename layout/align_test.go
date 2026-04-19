package layout

import (
	"fmt"
	"os"
	"slices"
	"testing"

	"github.com/redexp/familymarkup-lsp/state"
	"github.com/redexp/familymarkup-lsp/types"
)

func TestAlign(t *testing.T) {
	root := testRoot(t)

	list, relations := Align(root, AlignParams{
		FontRatio: 0.615,
	})

	if len(list) == 0 {
		t.Error("list == 0")
		return
	}

	if len(relations) == 0 {
		t.Error("relations == 0")
		return
	}

	var rects []cRect

	for _, family := range list {
		for li, level := range family.levels {
			for ri, rect := range level.Rects {
				rects = append(rects, cRect{
					Rect:  rect.Move(family.X, family.Y),
					color: "black",
					title: fmt.Sprintf("%s, l: %d, r: %d", family.Title.Name, li, ri),
				})
			}
		}
	}

	for _, family := range list {
		family.Walk(func(person *SvgPerson) {
			rects = append(rects, cRect{
				Rect:  person.Rect.Move(family.X, family.Y),
				color: "red",
				title: family.Title.Name + ", " + person.Name,
			})
		})
	}

	slices.SortFunc(rects, func(a, b cRect) int {
		return a.Y - b.Y
	})

	_RectsToSvg("after.svg", rects)
}

func testRoot(t *testing.T) *state.Root {
	root := state.CreateRoot()
	root.SetFolders([]types.Uri{"/home/sergii/projects/Родина"})
	err := root.UpdateDirty()

	if err != nil {
		t.Fatal(err)
	}

	return root
}

type cRect struct {
	Rect

	color string
	title string
}

func _RectsToSvg(filename string, rects []cRect) {
	minX := 0
	maxX := 0
	minY := 0
	maxY := 0
	for _, rect := range rects {
		right := rect.Right()
		bottom := rect.Y + rect.Height

		if rect.X < minX {
			minX = rect.X
		}
		if rect.Y < minY {
			minY = rect.Y
		}

		if maxX < right {
			maxX = right
		}
		if maxY < bottom {
			maxY = bottom
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	write := func(v string) {
		if _, err = file.WriteString(v); err != nil {
			panic(err)
		}
	}

	write(fmt.Sprintf(
		"<svg width=\"%d\" height=\"%d\" viewBox=\"%d %d %d %d\" xmlns=\"http://www.w3.org/2000/svg\">\n",
		maxX+10-minX, maxY+10-minY, minX, minY, maxX+10-minX, maxY+10-minY,
	))

	for _, r := range rects {
		write(fmt.Sprintf("\t<rect x='%d' y='%d' width='%d' height='%d' fill='%s' title='%s'/>\n", r.X, r.Y, r.Width, r.Height, r.color, r.title))
	}

	write("</svg>")
}

func _PointsToSVG(filename string, points []Pos) {
	maxX := 0
	maxY := 0
	for _, p := range points {
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	write := func(v string) {
		if _, err = file.WriteString(v); err != nil {
			panic(err)
		}
	}

	write(fmt.Sprintf(
		"<svg width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\" xmlns=\"http://www.w3.org/2000/svg\">\n",
		maxX+10, maxY+10, maxX+10, maxY+10,
	))

	write("<path stroke='black' stroke-width='2' fill='none' d='")

	for i, p := range points {
		if i == 0 {
			write(fmt.Sprintf("M%d,%d", p.X, p.Y))
			continue
		}

		write(fmt.Sprintf(" L%d,%d", p.X, p.Y))
	}

	write("'/>")
	write("</svg>")
}
