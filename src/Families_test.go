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

	file := filepath.Join(dir, "../test/root/simple.family")

	tree, text, err := getTreeText(file)

	if err != nil {
		t.Error(err)
		return
	}

	root := Families{}

	err = root.Update(tree, text, file)

	if err != nil {
		t.Error(err)
		return
	}
}
