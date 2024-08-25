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

	uri := toUri(filepath.Join(dir, "../test/root/simple.family"))

	doc, err := openDoc(uri)

	if err != nil {
		t.Error(err)
		return
	}

	root := createRoot()

	err = root.Update(doc.Tree, []byte(doc.Text), uri)

	if err != nil {
		t.Error(err)
		return
	}
}
