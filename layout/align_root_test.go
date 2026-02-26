package layout

import "testing"

func TestMergeLayouts(t *testing.T) {
	root := &AlignRoot{
		levels: []*Level{
			{
				Y: 0,
				Rects: []Rect{
					{X: 0, Width: 10},
					{X: 20, Width: 10},
				},
			},
		},
	}

	root.mergeLevels(
		Pos{X: -20, Y: -10},
		[]*Level{
			{
				Y: 0,
				Rects: []Rect{
					{X: 0, Width: 10},
				},
			},
			{
				Y: 10,
				Rects: []Rect{
					{X: 0, Width: 10},
				},
			},
			{
				Y: 30,
				Rects: []Rect{
					{X: 0, Width: 10},
				},
			},
		},
	)

	if len(root.levels) != 3 {
		t.Errorf("len(root.levels) != 3, %d", len(root.levels))
		return
	}

	level := root.levels[1]

	if len(level.Rects) != 1 {
		t.Errorf("len(level.Rects) != 1, %d", len(level.Rects))
		return
	}

	rect := level.Rects[0]

	if !(rect.X == -20 && rect.Width == 50) {
		t.Errorf("rect (-20, 50), %d, %d", rect.X, rect.Width)
		return
	}
}

func TestRootAlign(t *testing.T) {
	lh := ss.LevelHeight

	a := &AlignRoot{
		Pos: Pos{100, 100},
		levels: []*Level{
			{
				Y: 0,
				Rects: []Rect{
					{
						X:     0,
						Width: 60,
					},
				},
			},
			{
				Y: lh,
				Rects: []Rect{
					{
						X:     0,
						Width: 20,
					},
					{
						X:     40,
						Width: 20,
					},
				},
			},
		},
	}

	b := &AlignRoot{
		Pos: Pos{100, 100},
		levels: []*Level{
			{
				Y: 0,
				Rects: []Rect{
					{
						X:     20,
						Width: 20,
					},
				},
			},
			{
				Y: lh,
				Rects: []Rect{
					{
						X:     0,
						Width: 60,
					},
				},
			},
		},
	}

	from := Rect{
		Y:     lh,
		X:     40,
		Width: 20,
	}

	to := Rect{
		Y:     0,
		X:     20,
		Width: 20,
	}

	pos := a.align(
		b.levels,
		from,
		to,
	)

	if pos.X == 0 && pos.Y == 0 {
		t.Errorf("not found")
	}

	var rects []Rect

	for _, level := range a.levels {
		for _, rect := range level.Rects {
			rect.X += a.X
			rect.Y = level.Y + a.Y
			rect.Height = ss.LevelHeight
			rects = append(rects, rect)
		}
	}

	for _, level := range b.levels {
		for _, rect := range level.Rects {
			rect.X += pos.X + b.X
			rect.Y = level.Y + pos.Y + b.Y
			rect.Height = ss.LevelHeight
			rects = append(rects, rect)
		}
	}

	rects = append(rects, from.Move(a.X, a.Y))
	rects = append(rects, to.Move(b.X+pos.X, b.Y+pos.Y))
}
