package diff

func ratio(matches, length int) float64 {
	if length > 0 {
		return 2.0 * float64(matches) / float64(length)
	}
	return 1.0
}

func quickRatio(a, b []string) float64 {
	fullBCount := make(map[string]int)

	// viewing a and b as multisets, set matches to the cardinality
	// of their intersection; this counts the number of matches
	// without regard to order, so is clearly an upper bound
	for _, s := range b {
		fullBCount[s] = fullBCount[s] + 1
	}

	// avail[x] is the number of times x appears in 'b' less the
	// number of times we've seen it in 'a' so far ... kinda
	avail := make(map[string]int)
	matches := 0
	for _, s := range a {
		n, ok := avail[s]
		if !ok {
			n = fullBCount[s]
		}
		avail[s] = n - 1
		if n > 0 {
			matches++
		}
	}

	return ratio(matches, len(a)+len(b))
}

func realQuickRatio(a, b []string) float64 {
	la, lb := len(a), len(b)
	return ratio(min(la, lb), la+lb)
}
