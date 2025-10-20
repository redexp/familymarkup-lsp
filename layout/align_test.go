package layout

import (
	"testing"

	"github.com/redexp/familymarkup-lsp/state"
	"github.com/redexp/familymarkup-lsp/types"
)

func TestAlign(t *testing.T) {
	root := state.CreateRoot()
	root.SetFolders([]types.Uri{"/home/sergii/projects/relatives"})
	err := root.UpdateDirty()

	if err != nil {
		t.Error(err)
		return
	}

	list := Align(root, "file:///home/sergii/projects/relatives/Ключник/Ключник.family", AlignParams{
		FontRatio: 1,
	})

	if len(list) == 0 {
		t.Error("list == 0")
	}
}
