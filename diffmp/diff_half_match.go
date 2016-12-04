package diffmp

/**
 * Does a substring of shorttext exist within longtext such that the substring
 * is at least half the length of longtext?
 * @param {string} longtext Longer string.
 * @param {string} shorttext Shorter string.
 * @param {number} i Start index of quarter length substring within longtext.
 * @return {Array.<string>} Five element Array, containing the prefix of
 *     longtext, the suffix of longtext, the prefix of shorttext, the suffix
 *     of shorttext and the common middle.  Or null if there was no match.
 * @private
 */
func diffHalfMatchI(l, s []rune, i int) [][]rune {
	// Start with a 1/4 length substring at position i as a seed.
	seed := l[i : i+len(l)/4]
	j := -1
	common := []rune{}
	longA := []rune{}
	longB := []rune{}
	shortA := []rune{}
	shortB := []rune{}

	if j < len(s) {
		j = runesIndexOf(s, seed, j+1)
		for {
			if j == -1 {
				break
			}

			prefixLen := commonPrefixLen(l[i:], s[j:])
			suffixLen := commonSuffixLen(l[:i], s[:j])
			if len(common) < suffixLen+prefixLen {
				common = concat(s[j-suffixLen:j], s[j:j+prefixLen])
				longA = l[:i-suffixLen]
				longB = l[i+prefixLen:]
				shortA = s[:j-suffixLen]
				shortB = s[j+prefixLen:]
			}
			j = runesIndexOf(s, seed, j+1)
		}
	}

	if len(common)*2 >= len(l) {
		return [][]rune{
			longA, longB,
			shortA, shortB,
			common,
		}
	}
	return nil
}

func diffHalfMatch(dmp *DMP, text1, text2 []rune) [][]rune {
	if dmp.DiffTimeout <= 0 {
		// Don't risk returning a non-optimal diff if we have unlimited time.
		return nil
	}

	var long, short []rune
	if len(text1) > len(text2) {
		long = text1
		short = text2
	} else {
		long = text2
		short = text1
	}

	if len(long) < 4 || len(short)*2 < len(long) {
		return nil // Pointless.
	}

	// First check if the second quarter is the seed for a half-match.
	hm1 := diffHalfMatchI(long, short, int(float64(len(long)+3)/4))

	// Check again based on the third quarter.
	hm2 := diffHalfMatchI(long, short, int(float64(len(long)+1)/2))

	hm := [][]rune{}
	if hm1 == nil && hm2 == nil {
		return nil
	} else if hm2 == nil {
		hm = hm1
	} else if hm1 == nil {
		hm = hm2
	} else {
		// Both matched.  Select the longest.
		if len(hm1[4]) > len(hm2[4]) {
			hm = hm1
		} else {
			hm = hm2
		}
	}

	// A half-match was found, sort out the return data.
	if len(text1) > len(text2) {
		return hm
	}
	return [][]rune{hm[2], hm[3], hm[0], hm[1], hm[4]}
}
