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

func TestChange(t *testing.T) {
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
		doc := CreateDoc("", "Test\n\nName+Name")

		doc.Change(proto.TextDocumentContentChangeEvent{
			Range: &proto.Range{
				Start: item.Start,
				End:   item.End,
			},
			Text: item.Text,
		})

		if doc.Text != item.Test {
			t.Errorf("%d - got: %s; expect: %s", i+1, doc.Text, item.Test)
		}
	}
}
