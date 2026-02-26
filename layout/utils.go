package layout

func abs(i int) int {
	if i < 0 {
		return -i
	}

	return i
}

func lastItem[T any](s []T) T {
	return s[len(s)-1]
}
