package providers

import (
	"testing"

	. "github.com/redexp/familymarkup-lsp/state"
	. "github.com/redexp/familymarkup-lsp/types"
)

func TestSvgDocument(t *testing.T) {
	root = CreateRoot()
	root.SetFolders([]Uri{"/home/sergii/projects/relatives"})
	err := root.UpdateDirty()

	if err != nil {
		t.Error(err)
		return
	}

	list, err := SvgDocument(nil, &SvgDocumentParams{
		URI:       "file:///home/sergii/projects/relatives/Ключник/Ключник.family",
		FontRatio: 1,
	})

	if err != nil {
		t.Error(err)
		return
	}

	if len(list) == 0 {
		t.Error("len(list) == 0")
		return
	}

	f := list[0]

	if f == nil {
		return
	}
}
