package layout

import "slices"

type LevelsMap map[int][]Rect

type Level struct {
	Y      int
	Height int
	Rects  []Rect
}

func (lvlMap LevelsMap) Add(rect Rect) {
	lvlMap[rect.Y] = append(lvlMap[rect.Y], rect)
}

func (lvlMap LevelsMap) ToArray() (levels []*Level) {
	for y, rects := range lvlMap {
		first := rects[0]
		levels = append(levels, &Level{
			Y:      y,
			Height: first.Height,
			Rects:  rects,
		})
	}

	slices.SortFunc(levels, func(a, b *Level) int {
		return a.Y - b.Y
	})

	for i := 0; i < len(levels)-1; i++ {
		level := levels[i]
		next := levels[i+1]

		level.Height = next.Y - level.Y
	}

	return levels
}

func (lvlMap LevelsMap) Border(minGap int) []Pos {
	levels := lvlMap.ToArray()

	mergeLevelsRects(levels, minGap)

	return levelsToBounding(levels)
}

func (level *Level) Left() int {
	rect := level.Rects[0]

	return rect.X
}

func (level *Level) Right() int {
	rect := lastItem(level.Rects)

	return rect.Right()
}

func (level *Level) FindRect(rect Rect) *Rect {
	right := rect.Right()

	for _, r := range level.Rects {
		if r.Right() < rect.X || right < r.X {
			continue
		}

		return &r
	}

	return nil
}

func mergeLevelsRects(levels []*Level, minGap int) {
	for _, level := range levels {
		mergeRects(level, minGap)
	}

	for i := len(levels) - 2; i >= 0; i-- {
		level := levels[i]
		prevLevel := levels[i+1]
		list := make([]Rect, 1, len(level.Rects))
		list[0] = level.Rects[0]
		last := &list[0]

		for li := 1; li < len(level.Rects); li++ {
			next := level.Rects[li]
			start := last.Right()
			end := next.X

			for _, rect := range prevLevel.Rects {
				left := rect.X

				if left >= end {
					break
				}

				right := rect.Right()

				if right <= start {
					continue
				}

				if left <= start && end <= right {
					start = end
					break
				}

				if start <= left && left < end {
					end = left
				}

				if start < right && right <= end {
					start = right
				}
			}

			if end-start < minGap {
				last.Width = next.Right() - last.X
				continue
			}

			if start != last.Right() {
				last.Width = start - last.X
			}

			if end != next.X {
				next.Width = next.Right() - end
				next.X = end
			}

			list = append(list, next)
			last = &list[len(list)-1]
		}

		level.Rects = list
	}
}

func mergeRects(level *Level, minGap int) {
	slices.SortFunc(level.Rects, func(a, b Rect) int {
		return a.X - b.X
	})

	list := make([]Rect, 0, len(level.Rects))
	var last *Rect

	for _, rect := range level.Rects {
		if last != nil && rect.X-last.Right() < minGap {
			last.Width = rect.Right() - last.X
			continue
		}

		list = append(list, rect)
		last = &list[len(list)-1]
	}

	level.Rects = list
}

type LevelLine struct {
	Y      int
	Points []*BorderPoint
}

type BorderPoint struct {
	Pos

	Type string
	TypI int
	Prev *BorderPoint
	Next *BorderPoint
	Up   *BorderPoint
	Down *BorderPoint
}

func levelsToBounding(levels []*Level) (points []Pos) {
	list := levels
	levels = make([]*Level, len(list))
	copy(levels, list)

	levels[0] = &Level{
		Y:      ss.LevelHeight - int(ss.FamilyTitleSize),
		Height: int(ss.FamilyTitleSize),
		Rects:  list[0].Rects,
	}

	if levels[1].Left() < levels[0].Left() {
		levels[0].Height += ss.ArrowsHeight

		levels[1] = &Level{
			Y:      ss.LevelHeight + ss.ArrowsHeight,
			Height: ss.LevelHeight - ss.ArrowsHeight,
			Rects:  levels[1].Rects,
		}
	}

	lines := make([]*LevelLine, len(levels)+1)
	types := []string{TL, TR, BL, BR}
	rectPoints := make([]*BorderPoint, 4)
	count := 0

	lines[0] = &LevelLine{}

	for i, level := range levels {
		lines[i+1] = &LevelLine{}

		line := lines[i]
		next := lines[i+1]

		line.Y = level.Y
		next.Y = line.Y + level.Height

		for _, rect := range level.Rects {
			rect.Y = level.Y
			rect.Height = level.Height

			for ti, t := range types {
				rectPoints[ti] = &BorderPoint{
					Pos:  rect.ToPos(t),
					Type: t,
					TypI: ti,
				}
			}

			line.Points = append(
				line.Points,
				rectPoints[0],
				rectPoints[1],
			)

			next.Points = append(
				next.Points,
				rectPoints[2],
				rectPoints[3],
			)

			rectPoints[0].Down = rectPoints[2]
			rectPoints[1].Down = rectPoints[3]
			rectPoints[2].Up = rectPoints[0]
			rectPoints[3].Up = rectPoints[1]

			count += 4
		}
	}

	for _, line := range lines {
		slices.SortFunc(line.Points, func(a, b *BorderPoint) int {
			if a.X == b.X {
				return a.TypI - b.TypI
			}

			return a.X - b.X
		})

		var prev *BorderPoint

		for _, p := range line.Points {
			p.Prev = prev

			if prev != nil {
				prev.Next = p
			}

			prev = p
		}
	}

	point := lines[0].Points[0]
	var prev *Pos

	for i := 0; i < count; i++ {
		if prev == nil || *prev != point.Pos {
			points = append(points, point.Pos)
			prev = &point.Pos
		}

		switch point.Type {
		case TL:
			point = point.Down
		case BR:
			point = point.Up
		case BL:
			if point.Prev != nil && point.Prev.Type == TL {
				point = point.Prev
			} else {
				point = point.Next
			}
		case TR:
			if point.Next != nil {
				point = point.Next
			} else {
				point = point.Prev
			}
		}
	}

	return addBorderPadding(points)
}

func addBorderPadding(points []Pos) []Pos {
	last := len(points) - 1
	list := make([]Pos, last+1)
	p := ss.BorderPadding

	list[0] = points[0].Move(-p, -p)
	list[last] = points[last].Move(p, -p)

	for i := 1; i < last; i++ {
		prev := points[i-1]
		cur := points[i]
		next := points[i+1]
		x := 0
		y := 0

		if prev == cur {
			list[i] = list[i-1]
			continue
		}

		if prev.X == cur.X && prev.Y < cur.Y {
			x = -1
			if next.X < cur.X {
				// ↖
				y = -1
			} else if next.X > cur.X {
				// ↙
				y = 1
			}
		} else if prev.Y == cur.Y && cur.X < prev.X {
			y = -1
			if next.Y < cur.Y {
				// ↗
				x = 1
			} else if next.Y > cur.Y {
				// ↖
				x = -1
			}
		} else if prev.Y == cur.Y && prev.X < cur.X {
			y = 1
			if next.Y < cur.Y {
				// ↘
				x = 1
			} else if next.Y > cur.Y {
				// ↙
				x = -1
			}
		} else if prev.X == cur.X && prev.Y > cur.Y {
			x = 1
			if next.X < cur.X {
				// ↗
				y = -1
			} else if next.X > cur.X {
				// ↘
				y = 1
			}
		}

		if y == -1 && i > 2 && i < last-2 {
			y = 0
		}

		list[i] = cur.Move(p*x, p*y)
	}

	return list
}
