package providers

import (
	"testing"

	familymarkup "github.com/redexp/tree-sitter-familymarkup"
)

func TestLegend(t *testing.T) {
	legend, err := familymarkup.GetHighlightLegend()

	if err != nil {
		t.Error(err)
		return
	}

	l, types, err := GetLegend()

	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%v, %v, %v", legend, l, types)
}
