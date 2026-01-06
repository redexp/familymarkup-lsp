package layout

import (
	"testing"
)

func TestMergeLevelsRects(t *testing.T) {
	levels := []Level{
		{
			Rects: []Rect{
				{
					X:     10,
					Width: 10,
				},
				{
					X:     30,
					Width: 30,
				},
			},
		},
		{
			Rects: []Rect{
				{
					X:     0,
					Width: 20,
				},
				{
					X:     40,
					Width: 30,
				},
			},
		},
		{
			Rects: []Rect{
				{
					X:     10,
					Width: 20,
				},
			},
		},
	}

	result := []Level{
		{
			Rects: []Rect{
				{
					X:     10,
					Width: 50,
				},
			},
		},
		{
			Rects: []Rect{
				{
					X:     0,
					Width: 30,
				},
				{
					X:     40,
					Width: 30,
				},
			},
		},
		{
			Rects: []Rect{
				{
					X:     10,
					Width: 20,
				},
			},
		},
	}

	MergeLevelsRects(levels, 10)

	for i, level := range levels {
		resLevel := result[i]

		if len(level.Rects) != len(resLevel.Rects) {
			t.Errorf("level.Rects = %d, resLevel.Rects = %d \n", len(level.Rects), len(resLevel.Rects))
			continue
		}

		for j, rect := range level.Rects {
			resRect := resLevel.Rects[j]

			if rect != resRect {
				t.Errorf("level %d: rect (%d, %d) != resRect (%d, %d)", i, rect.X, rect.Width, resRect.X, resRect.Width)
			}
		}
	}
}

func TestLevelsBorder(t *testing.T) {
	levels := []Level{
		{
			Y:      0,
			Height: 10,
			Rects: []Rect{
				{
					X:     10,
					Width: 10,
				},
				{
					X:     30,
					Width: 30,
				},
			},
		},
		{
			Y:      10,
			Height: 10,
			Rects: []Rect{
				{
					X:     0,
					Width: 30,
				},
				{
					X:     40,
					Width: 30,
				},
			},
		},
		{
			Y:      20,
			Height: 10,
			Rects: []Rect{
				{
					X:     10,
					Width: 20,
				},
			},
		},
	}

	MergeLevelsRects(levels, 20)

	result := LevelsBorder(levels)

	points := []Pos{
		{10, 0},
		{10, 10},
		{0, 10},
		{0, 20},
		{10, 20},
		{10, 30},
		{30, 30},
		{30, 20},
		{70, 20},
		{70, 10},
		{60, 10},
		{60, 0},
	}

	if len(result) != len(points) {
		t.Errorf("result %d != points %d \n", len(result), len(points))
		return
	}

	for i, r := range result {
		p := points[i]

		if r != p {
			t.Errorf("i = %d, p = %v, r = %v \n", i, p, r)
		}
	}
}
