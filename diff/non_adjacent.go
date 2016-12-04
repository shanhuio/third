package diff

func nonAdjacent(matched []Match, na, nb int) []Match {
	// It's possible that we have adjacent equal blocks in the
	// matching_blocks list now.
	var ret []Match
	i1, j1, k1 := 0, 0, 0
	for _, b := range matched {
		// Is this block adjacent to i1, j1, k1?
		i2, j2, k2 := b.A, b.B, b.Size
		if i1+k1 == i2 && j1+k1 == j2 {
			// Yes, so collapse them -- this just increases the length of
			// the first block by the length of the second, and the first
			// block so lengthened remains the block to compare against.
			k1 += k2
		} else {
			// Not adjacent.  Remember the first block (k1==0 means it's
			// the dummy we started with), and make the second block the
			// new block to compare against.
			if k1 > 0 {
				ret = append(ret, Match{i1, j1, k1})
			}
			i1, j1, k1 = i2, j2, k2
		}
	}

	if k1 > 0 {
		ret = append(ret, Match{i1, j1, k1})
	}

	return append(ret, Match{na, nb, 0})
}
