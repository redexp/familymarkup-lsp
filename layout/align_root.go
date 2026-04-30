package layout

import (
	"slices"
)

type AlignRoot struct {
	Pos

	levels   []*Level
	families []*SvgFamily
}

func alignByLevels(families []*SvgFamily) {
	if len(families) <= 1 {
		return
	}

	moved := make(map[*SvgFamily]*AlignRoot)
	linksSet := make(map[*SvgFamilyLink]struct{})

	move := func(f *SvgFamily, links []*SvgFamilyLink) {
		root := moved[f]

		if root == nil {
			root = &AlignRoot{
				families: []*SvgFamily{f},
			}

			root.mergeLevels(Pos{}, f.levels)

			moved[f] = root
		}

		for _, link := range links {
			if _, exist := linksSet[link]; exist {
				continue
			}

			linksSet[link] = struct{}{}

			target := link.Family
			from := link.From.Move(f.X, f.Y)
			to := link.To.Move(target.X, target.Y)

			if targetRoot, exist := moved[target]; exist {
				if root == targetRoot {
					continue
				}

				pos := root.align(targetRoot.levels, from, to)
				root.mergeRoot(pos, targetRoot)

				for _, family := range targetRoot.families {
					moved[family] = root
				}

				continue
			}

			moved[target] = root
			root.families = append(root.families, target)

			pos := root.align(target.levels, from, to)

			target.X = pos.X
			target.Y = pos.Y

			root.mergeLevels(pos, target.levels)
		}
	}

	var cluster []*SvgFamily

	flush := func() {
		if len(cluster) > 1 {
			for _, f := range cluster {
				links := make([]*SvgFamilyLink, 0, len(f.links))
				for _, link := range f.links {
					if slices.Contains(cluster, link.Family) {
						links = append(links, link)
					}
				}

				move(f, links)
			}
		}

		cluster = make([]*SvgFamily, 0)
	}

	prevUri := families[0].Uri

	for _, f := range families {
		if f.Uri == prevUri {
			cluster = append(cluster, f)
			continue
		}

		flush()
		cluster = append(cluster, f)
		prevUri = f.Uri
	}

	flush()

	for _, f := range families {
		move(f, f.links)
	}

	rootMap := make(map[*AlignRoot]struct{})
	prev := moved[families[0]]
	rootMap[prev] = struct{}{}

	for _, f := range families {
		root := moved[f]

		if _, ok := rootMap[root]; ok {
			continue
		}

		rootMap[root] = struct{}{}

		prevRect := prev.Rect()
		y := prevRect.Y + prevRect.Height
		rootRect := root.Rect()
		top := rootRect.Y

		if top < 0 {
			y += -top
		}

		x := (prevRect.X + prevRect.Width/2) - rootRect.Width/2

		root.Move(x, y)

		prev = root
	}
}

func (root *AlignRoot) Move(x, y int) {
	root.X += x
	root.Y += y

	for _, f := range root.families {
		f.X += x
		f.Y += y
	}
}

func (root *AlignRoot) Height() int {
	return len(root.levels) * ss.LevelHeight
}

func (root *AlignRoot) Top() int {
	return root.levels[0].Y + root.Y
}

func (root *AlignRoot) Bottom() int {
	last := lastItem(root.levels)

	return last.Y + ss.LevelHeight + root.Y
}

func (root *AlignRoot) Rect() Rect {
	y := root.Top()
	height := root.Height()

	minLeft := root.levels[0].Left()
	maxRight := root.levels[0].Right()

	for _, level := range root.levels {
		left := level.Left()
		right := level.Right()

		if left < minLeft {
			minLeft = left
		}

		if right > maxRight {
			maxRight = right
		}
	}

	return Rect{
		X:      root.X + minLeft,
		Y:      y,
		Width:  maxRight - minLeft,
		Height: height,
	}
}

func (root *AlignRoot) align(levels []*Level, from, to Rect) Pos {
	type Result struct {
		Pos
		dir      int
		distance int
	}

	fromLevel, fromLevelRect := findLevelByRect(root.levels, from)
	toLevel, _ := findLevelByRect(levels, to)

	fromRect := Rect{
		X:      from.X,
		Y:      fromLevel.Y,
		Width:  from.Width,
		Height: fromLevel.Height,
	}
	fromPos := fromRect.ToPos("center")

	getDistance := func(pos *Result) int {
		toRect := Rect{
			X:      to.X + pos.X,
			Y:      toLevel.Y + pos.Y,
			Width:  to.Width,
			Height: toLevel.Height,
		}

		toPos := toRect.ToPos("center")

		x := abs(fromPos.X - toPos.X)
		y := abs(fromPos.Y - toPos.Y)

		return x + y
	}

	results := []*Result{
		{
			Pos: Pos{
				X: from.X - to.X,
				Y: fromLevel.Y + ss.LevelHeight,
			},
			dir: 0,
		},
	}

	results[0].distance = getDistance(results[0])

	rootFirstLevel := root.levels[0]
	rootLastLevel := lastItem(root.levels)

	targetFirstLevel := levels[0]
	targetLastLevel := lastItem(levels)

	topStart := -(targetLastLevel.Y + ss.LevelHeight - rootFirstLevel.Y)

	for targetFirstLevel.Y+topStart < rootLastLevel.Y {
		res := &Result{}
		res.dir = 0
		res.X = results[0].X
		res.Y = targetFirstLevel.Y + topStart
		res.distance = getDistance(res)
		results = append(results, res)
		topStart += ss.LevelHeight
	}

	for _, level := range levels {
		for _, rect := range level.Rects {
			res := &Result{}
			res.X = fromLevelRect.X - rect.Right()
			res.Y = fromLevel.Y - level.Y
			res.dir = -1
			res.distance = getDistance(res)

			results = append(results, res)

			res = &Result{}
			res.X = fromLevelRect.Right() - rect.X
			res.Y = fromLevel.Y - level.Y
			res.dir = 1
			res.distance = getDistance(res)

			results = append(results, res)
		}
	}

	sort := func() {
		slices.SortFunc(results, func(a, b *Result) int {
			if a == nil {
				if b == nil {
					return 0
				}

				return 1
			}

			if b == nil {
				return -1
			}

			return a.distance - b.distance
		})
	}

	sort()

	defRoot := &AlignRoot{
		Pos:    Pos{0, 0},
		levels: root.levels,
	}

	for {
		res := results[0]

		if res == nil {
			panic("can't find levels position")
		}

		pos := res.Pos

		posRoot := &AlignRoot{
			Pos:    pos,
			levels: levels,
		}

		diff := posRoot.getLevelsIntersection(defRoot)

		if diff == 0 {
			pos.X += ss.BorderPadding * 2 * res.dir

			return pos
		}

		if res.dir == 0 {
			results[0] = nil
			sort()
			continue
		}

		res.X += diff * res.dir

		res.distance = getDistance(res)

		sort()
	}
}

func (root *AlignRoot) mergeLevels(pos Pos, levels []*Level) {
	relX := pos.X - root.X
	relY := pos.Y - root.Y

	list := make([]*Level, len(levels))

	for i, level := range levels {
		item := &Level{
			Y:      level.Y + relY,
			Height: level.Height,
			Rects:  make([]Rect, len(level.Rects)),
		}

		for ri, rect := range level.Rects {
			rect.X += relX
			rect.Y = item.Y
			item.Rects[ri] = rect
		}

		list[i] = item
	}

	if len(root.levels) == 0 {
		root.levels = list
		return
	}

	levels = list

	lvlMap := make(map[int]*Level)

	for _, level := range root.levels {
		lvlMap[level.Y] = level
	}

	before := make([]*Level, 0, len(levels))
	after := make([]*Level, 0, len(levels))
	firstLevel := root.levels[0]

	for _, level := range levels {
		if level.Y < firstLevel.Y {
			before = append(before, level)
			continue
		}

		curLevel := lvlMap[level.Y]

		if curLevel == nil {
			after = append(after, level)
			continue
		}

		if level.Rects[0].X < curLevel.Rects[0].X {
			curLevel.Rects = append(level.Rects, curLevel.Rects...)
		} else {
			curLevel.Rects = append(curLevel.Rects, level.Rects...)
		}

		mergeRects(curLevel, ss.LevelMinGap)
	}

	if len(before) > 0 {
		root.levels = append(before, root.levels...)
	}

	if len(after) > 0 {
		root.levels = append(root.levels, after...)
	}
}

func (root *AlignRoot) mergeRoot(pos Pos, source *AlignRoot) {
	for _, f := range source.families {
		f.X += pos.X
		f.Y += pos.Y
	}

	root.families = append(root.families, source.families...)

	root.mergeLevels(pos, source.levels)
}

func (root *AlignRoot) getLevelsIntersection(target *AlignRoot) int {
	ao := 0
	bo := 0
	rTop := root.Top()
	tTop := target.Top()

	if rTop < tTop {
		ao = (tTop - rTop) / ss.LevelHeight
	} else if tTop < rTop {
		bo = (rTop - tTop) / ss.LevelHeight
	}

	aLen := len(root.levels)
	bLen := len(target.levels)
	maxLen := max(aLen, bLen)

	for i := 0; i < maxLen; i++ {
		ai := i + ao
		bi := i + bo

		if ai == aLen || bi == bLen {
			return 0
		}

		aLevel := root.levels[ai]
		bLevel := target.levels[bi]

		diff := getIntervalsDiff(root.X, aLevel, target.X, bLevel)

		if diff > 0 {
			return diff
		}
	}

	return 0
}

func findLevelByRect(levels []*Level, rect Rect) (level *Level, lRect *Rect) {
	for i, l := range levels {
		if rect.Y == l.Y {
			level = l
			break
		}

		if rect.Y < l.Y {
			if i == 0 {
				panic("i == 0")
			}

			level = levels[i-1]
			break
		}
	}

	if level == nil {
		level = lastItem(levels)
	}

	for i, r := range level.Rects {
		if rect.X == r.X {
			lRect = &r
			break
		}

		if rect.X < r.X {
			if i == 0 {
				panic("empty rects")
			}

			lRect = &level.Rects[i-1]
			break
		}
	}

	if lRect == nil {
		if len(level.Rects) == 0 {
			panic("empty rects")
		}

		lRect = &level.Rects[len(level.Rects)-1]
	}

	return
}

func getIntervalsDiff(ax int, al *Level, bx int, bl *Level) int {
	i, j := 0, 0

	for i < len(al.Rects) && j < len(bl.Rects) {
		segA := al.Rects[i]
		segB := bl.Rects[j]

		aRight := segA.Right() + ax
		bRight := segB.Right() + bx

		overlapStart := max(segA.X+ax, segB.X+bx)
		overlapEnd := min(aRight, bRight)

		if overlapStart < overlapEnd {
			return overlapEnd - overlapStart
		}

		if aRight < bRight {
			i++
		} else {
			j++
		}
	}

	return 0
}
