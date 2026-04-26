package layout

import (
	"testing"
)

func TestGraph(t *testing.T) {
	root := testRoot(t)

	families, relations := CreateGraphFamilies(root)

	if len(families) == 0 {
		t.Error("len(families) == 0")
		return
	}

	if len(relations) == 0 {
		t.Error("len(relations) == 0")
		return
	}

	for n := 0; n < 10; n++ {
		list, _ := CreateGraphFamilies(testRoot(t))

		for i, f := range families {
			item := list[i]

			if item.Uri != f.Uri {
				t.Errorf("[%d] \n\titem.Uri = %s\n\tf.Uri = %s\n", i, item.Uri, f.Uri)
				return
			}

			if item.Name.Text != f.Name.Text {
				t.Errorf("[%d] \n\titem.Name = %s\n\tf.Name = %s\n", i, item.Name.Text, f.Name.Text)
				return
			}

			if item.Name.Line != f.Name.Line {
				t.Errorf("[%d] \n\titem.Line = %d\n\tf.Line = %d\n", i, item.Name.Line, f.Name.Line)
				return
			}

			var fPersons []*GraphPerson

			f.Walk(func(p *GraphPerson) {
				fPersons = append(fPersons, p)
			})

			var itemPersons []*GraphPerson

			item.Walk(func(p *GraphPerson) {
				itemPersons = append(itemPersons, p)
			})

			if len(itemPersons) != len(fPersons) {
				t.Errorf("itemPersons != fPersons")
				return
			}

			for j, fp := range fPersons {
				ip := itemPersons[j]
				//token := ip.Token()

				//if token != nil && token.Text == "Уляна" && f.Name.Text == "Савіч" {
				//	if ip.Link == nil && fp.Link != nil {
				//		t.Errorf("ip.Link == nil && fp.Link != nil")
				//		return
				//	}
				//}

				if ip.Person.Loc != fp.Person.Loc {
					t.Errorf("ip.Loc != fp.Loc")
					return
				}

				if ip.Link == nil && fp.Link != nil {
					t.Errorf("ip.Link == nil && fp.Link != nil")
					return
				}

				if ip.Link != nil && fp.Link == nil {
					t.Errorf("ip.Link != nil && fp.Link == nil")
					return
				}

				if ip.Link != nil && ip.Link.Person.Loc != fp.Link.Person.Loc {
					t.Errorf("ip.Link != nil && fp.Link == nil")
					return
				}
			}
		}
	}
}
