package src

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdate(t *testing.T) {
	dir, err := os.Getwd()

	if err != nil {
		t.Error(err)
		return
	}

	tree, text, err := getTreeText(filepath.Join(dir, "../test/root/simple.family"))

	if err != nil {
		t.Error(err)
		return
	}

	root := Families{}

	err = root.Update(tree, text)

	if err != nil {
		t.Error(err)
		return
	}
}
