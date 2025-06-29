package providers

import (
	fm "github.com/redexp/familymarkup-parser"
	proto "github.com/tliron/glsp/protocol_3_16"
)

func FoldingRange(_ *Ctx, params *proto.FoldingRangeParams) (res []proto.FoldingRange, err error) {
	uri := NormalizeUri(params.TextDocument.URI)

	doc := GetDoc(uri)

	res = getFoldingRanges(doc.Root)

	return
}

func getFoldingRanges(root *fm.Root) (list []proto.FoldingRange) {
	kind := P("region")

	add := func(loc fm.Loc) {
		list = append(list, proto.FoldingRange{
			StartLine: proto.UInteger(loc.Start.Line),
			EndLine:   proto.UInteger(loc.End.Line),
			Kind:      kind,
		})
	}

	for _, family := range root.Families {
		add(family.Loc)

		for _, rel := range family.Relations {
			add(rel.Loc)
		}
	}

	return
}
