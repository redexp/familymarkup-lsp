package layout

import (
	"fmt"
	"os"
	"testing"

	"github.com/redexp/familymarkup-lsp/state"
	"github.com/redexp/familymarkup-lsp/types"
)

func TestAlign(t *testing.T) {
	root := testRoot(t)

	list := Align(root, "file:///home/sergii/projects/Родина/Ключник/Ключник.family", AlignParams{
		FontRatio: 0.615,
	})

	if len(list) == 0 {
		t.Error("list == 0")
		return
	}
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

func _RectsToSvg(filename string, rects map[string][]Rect) {
	minX := 0
	maxX := 0
	minY := 0
	maxY := 0
	for _, list := range rects {
		for _, rect := range list {
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

	for color, list := range rects {
		for _, r := range list {
			write(fmt.Sprintf("\t<rect x='%d' y='%d' width='%d' height='%d' fill='%s'/>\n", r.X, r.Y, r.Width, r.Height, color))
		}
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
