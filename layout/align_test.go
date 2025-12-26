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

	list := Align(root, "file:///home/sergii/projects/relatives/Ключник/Ключник.family", AlignParams{
		FontRatio: 1,
	})

	if len(list) == 0 {
		t.Error("list == 0")
	}
}

func testRoot(t *testing.T) *state.Root {
	root := state.CreateRoot()
	root.SetFolders([]types.Uri{"/home/sergii/projects/relatives"})
	err := root.UpdateDirty()

	if err != nil {
		t.Fatal(err)
	}

	return root
}

func _RectsToSVG(filename string, points []Pos) (err error) {
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
		return err
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

	return nil
}
