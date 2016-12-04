package diffmp

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
