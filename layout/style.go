package layout

var ss = Style{
	FamilyTitleSize: 16,
	BorderPadding:   10,
	FamilyGap:       15,
	PersonNameSize:  12,
	PersonHeight:    30,
	PersonPaddingX:  20,
	PersonMarginX:   10,
	ArrowsHeight:    25,
	GridStep:        30,
	LevelHeight:     30 + 25,
	LevelMinGap:     160,
}

type Style struct {
	FamilyTitleSize float64
	BorderPadding   int
	FamilyGap       int
	PersonNameSize  float64
	PersonHeight    float64
	PersonPaddingX  float64
	PersonMarginX   float64
	ArrowsHeight    int
	GridStep        int
	LevelHeight     int
	LevelMinGap     int
}
