package diffmp

func concat(r1, r2 []rune) []rune {
	result := make([]rune, len(r1)+len(r2))
	copy(result, r1)
	copy(result[len(r1):], r2)
	return result
}
