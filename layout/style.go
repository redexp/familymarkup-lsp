package layout

var ss = Style{
	FamilyTitleSize: 16,
	FamilyPadding:   10,
	FamilyGap:       15,
	PersonNameSize:  12,
	PersonHeight:    30,
	PersonPaddingX:  20,
	PersonMarginX:   10,
	ArrowsHeight:    25,
	GridStep:        30,
}

type Style struct {
	FamilyTitleSize float64
	FamilyPadding   int
	FamilyGap       int
	PersonNameSize  float64
	PersonHeight    float64
	PersonPaddingX  float64
	PersonMarginX   float64
	ArrowsHeight    int
	GridStep        int
}
