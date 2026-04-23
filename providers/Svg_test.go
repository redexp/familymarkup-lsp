package providers

import (
	"testing"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
)

func TestSvgFamilies(t *testing.T) {
	root = testRoot(t)

	res, err := SvgFamilies(nil, &SvgFamiliesParams{
		URI:       "file:///home/sergii/projects/relatives/Ключник/Ключник.family",
		FontRatio: 1,
	})

	if err != nil {
		t.Error(err)
		return
	}

	if len(res.Families) == 0 {
		t.Error("len(res.Families) == 0")
		return
	}

	f := res.Families[0]

	if f == nil {
		return
	}
}

func TestSvgPath(t *testing.T) {
	root = testRoot(t)

	res, err := SvgPath(nil, &SvgPathParams{
		Persons: []SvgPathPerson{
			{
				URI: "file:///home/sergii/projects/Родина/Нагорні/Ивановы.family",
				Position: Position{
					Line:      16,
					Character: 3,
				},
			},
			{
				URI: "file:///home/sergii/projects/Родина/Ключник/Ключник.family",
				Position: Position{
					Line:      18,
					Character: 3,
				},
			},
		},
	})

	if err != nil {
		t.Error(err)
		return
	}

	if len(res.Path) == 0 {
		t.Errorf("len(res.Path) == 0")
		return
	}
}

func testRoot(t *testing.T) *Root {
	root := CreateRoot()
	root.SetFolders([]Uri{"/home/sergii/projects/Родина"})
	err := root.UpdateDirty()

	if err != nil {
		t.Fatal(err)
	}

	return root
}
