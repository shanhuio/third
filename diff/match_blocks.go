package diff

func matchBlocks(
	m *matcher, alo, ahi, blo, bhi int, matched []Match,
) []Match {
	match := findLongestMatch(m, alo, ahi, blo, bhi)
	i, j, k := match.A, match.B, match.Size
	if match.Size > 0 {
		if alo < i && blo < j {
			matched = matchBlocks(m, alo, i, blo, j, matched)
		}
		matched = append(matched, match)
		if i+k < ahi && j+k < bhi {
			matched = matchBlocks(m, i+k, ahi, j+k, bhi, matched)
		}
	}
	return matched
}
