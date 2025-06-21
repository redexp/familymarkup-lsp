package state

import (
	proto "github.com/tliron/glsp/protocol_3_16"
	"testing"
)

func TestDirtyUris_ChangeText(t *testing.T) {
	list := []struct {
		Start proto.Position
		End   proto.Position
		Text  string
		Test  string
	}{
		{
			Start: proto.Position{0, 0},
			End:   proto.Position{0, 4},
			Text:  "Fam",
			Test:  "Fam\n\nName+Name",
		},
		{
			Start: proto.Position{0, 0},
			End:   proto.Position{0, 5},
			Text:  "",
			Test:  "\nName+Name",
		},
		{
			Start: proto.Position{0, 2},
			End:   proto.Position{2, 1},
			Text:  "",
			Test:  "Teame+Name",
		},
		{
			Start: proto.Position{0, 2},
			End:   proto.Position{2, 2},
			Text:  "+",
			Test:  "Te+me+Name",
		},
		{
			Start: proto.Position{0, 0},
			End:   proto.Position{0, 0},
			Text:  "Fam-",
			Test:  "Fam-Test\n\nName+Name",
		},
		{
			Start: proto.Position{3, 0},
			End:   proto.Position{3, 0},
			Text:  "\nName",
			Test:  "Test\n\nName+Name\nName",
		},
		{
			Start: proto.Position{2, 1},
			End:   proto.Position{2, 1},
			Text:  "tt",
			Test:  "Test\n\nNttame+Name",
		},
	}

	for i, item := range list {
		doc := CreateDoc("/test.txt", "Test\n\nName+Name")
		uris := DirtyUris{}

		uris.ChangeText(
			doc,
			&proto.Range{
				Start: item.Start,
				End:   item.End,
			},
			item.Text,
		)

		text := uris[doc.Uri].Text

		if text != item.Test {
			t.Errorf("%d - got: %s; expect: %s", i+1, text, item.Test)
		}
	}
}
