package state

import (
	proto "github.com/tliron/glsp/protocol_3_16"
	"testing"
)

func TestGetTextByRange(t *testing.T) {
	doc := CreateDoc("", "Test\n\nName+Name")

	list := []struct {
		proto.Range
		Text string
	}{
		{
			Range: proto.Range{
				Start: proto.Position{0, 0},
				End:   proto.Position{2, 0},
			},
			Text: "Test\n\n",
		},
		{
			Range: proto.Range{
				Start: proto.Position{0, 0},
				End:   proto.Position{2, 4},
			},
			Text: "Test\n\nName",
		},
		{
			Range: proto.Range{
				Start: proto.Position{0, 2},
				End:   proto.Position{2, 3},
			},
			Text: "st\n\nNam",
		},
		{
			Range: proto.Range{
				Start: proto.Position{2, 1},
				End:   proto.Position{2, 3},
			},
			Text: "am",
		},
		{
			Range: proto.Range{
				Start: proto.Position{2, 1},
				End:   proto.Position{2, 5},
			},
			Text: "ame+",
		},
		{
			Range: proto.Range{
				Start: proto.Position{1, 0},
				End:   proto.Position{2, 4},
			},
			Text: "\nName",
		},
		{
			Range: proto.Range{
				Start: proto.Position{0, 0},
				End:   proto.Position{0, 10},
			},
			Text: "Test\n",
		},
		{
			Range: proto.Range{
				Start: proto.Position{0, 10},
				End:   proto.Position{2, 0},
			},
			Text: "\n",
		},
		{
			Range: proto.Range{
				Start: proto.Position{4, 0},
				End:   proto.Position{5, 0},
			},
			Text: "",
		},
		{
			Range: proto.Range{
				Start: proto.Position{0, 0},
				End:   proto.Position{0, 0},
			},
			Text: "",
		},
	}

	for i, item := range list {
		text := doc.GetTextByRange(item.Range)

		if text != item.Text {
			t.Errorf("%d - got: %s; expect: %s", i+1, text, item.Text)
		}
	}
}
