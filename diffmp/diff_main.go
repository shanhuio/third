package diffmp

import (
	"time"
)

func diffMain(
	dmp *DMP, s1, s2 string, checkLines bool, deadline time.Time,
) []Diff {
	return diffMainRunes(dmp, []rune(s1), []rune(s2), checkLines, deadline)
}

func diffMainRunes(
	dmp *DMP, s1, s2 []rune, checkLines bool, deadline time.Time,
) []Diff {
	if runesEqual(s1, s2) {
		var diffs []Diff
		if len(s1) > 0 {
			diffs = append(diffs, Diff{Noop, string(s1)})
		}
		return diffs
	}
	// Trim off common prefix (speedup).
	n := commonPrefixLen(s1, s2)
	prefix := s1[:n]
	s1 = s1[n:]
	s2 = s2[n:]

	// Trim off common suffix (speedup).
	n = commonSuffixLen(s1, s2)
	suffix := s1[len(s1)-n:]
	s1 = s1[:len(s1)-n]
	s2 = s2[:len(s2)-n]

	// Compute the diff on the middle block.
	diffs := diffCompute(dmp, s1, s2, checkLines, deadline)

	// Restore the prefix and suffix.
	if len(prefix) != 0 {
		diffs = diffPrepend(diffNoop(string(prefix)), diffs)
	}
	if len(suffix) != 0 {
		diffs = diffAppend(diffs, diffNoop(string(suffix)))
	}
	return DiffCleanupMerge(diffs)
}

// diffCompute finds the differences between two rune slices.  Assumes that
// the texts do not have any common prefix or suffix.
func diffCompute(
	dmp *DMP, s1, s2 []rune, checkLines bool, deadline time.Time,
) []Diff {
	diffs := []Diff{}
	if len(s1) == 0 {
		// Just add some text (speedup).
		return append(diffs, Diff{Insert, string(s2)})
	}
	if len(s2) == 0 {
		// Just delete some text (speedup).
		return append(diffs, Diff{Delete, string(s1)})
	}

	long, short := s1, s2
	if len(long) < len(short) {
		long, short = short, long
	}

	if i := runesIndex(long, short); i != -1 {
		op := Insert
		// Swap insertions for deletions if diff is reversed.
		if len(s1) > len(s2) {
			op = Delete
		}
		// Shorter text is inside the longer text (speedup).
		return []Diff{
			{op, string(long[:i])},
			diffNoop(string(short)),
			{op, string(long[i+len(short):])},
		}
	} else if len(short) == 1 {
		// Single character string.
		// After the previous speedup, the character can't be an equality.
		return []Diff{diffDel(string(s1)), diffIns(string(s2))}
		// Check to see if the problem can be split in two.
	} else if hm := diffHalfMatch(dmp, s1, s2); hm != nil {
		// A half-match was found, sort out the return data.
		s1a, s1b := hm[0], hm[1]
		s2a, s2b := hm[2], hm[3]
		midCommon := hm[4]
		// Send both pairs off for separate processing.
		diffsA := diffMainRunes(dmp, s1a, s2a, checkLines, deadline)
		diffsB := diffMainRunes(dmp, s1b, s2b, checkLines, deadline)
		// Merge the results.
		return append(diffsA,
			diffPrepend(diffNoop(string(midCommon)), diffsB)...,
		)
	} else if checkLines && len(s1) > 100 && len(s2) > 100 {
		return dmp.diffLineMode(s1, s2, deadline)
	}
	return diffBisect(dmp, s1, s2, deadline)
}

// diffLineMode does a quick line-level diff on both []runes, then rediff the
// parts for greater accuracy. This speedup can produce non-minimal diffs.
func (dmp *DMP) diffLineMode(s1, s2 []rune, deadline time.Time) []Diff {
	// Scan the text on a line-by-line basis first.
	s1, s2, linearray := diffLinesToRunes(s1, s2)
	diffs := diffMainRunes(dmp, s1, s2, false, deadline)

	// Convert the diff back to original text.
	diffs = DiffCharsToLines(diffs, linearray)
	// Eliminate freak matches (e.g. blank lines)
	diffs = DiffCleanupSemantic(diffs)

	// Rediff any replacement blocks, this time character-by-character.
	// Add a dummy entry at the end.
	diffs = append(diffs, Diff{Noop, ""})

	i := 0
	ndel, nins := 0, 0
	delStr, insStr := "", ""

	for i < len(diffs) {
		switch diffs[i].Type {
		case Insert:
			nins++
			insStr += diffs[i].Text
		case Delete:
			ndel++
			delStr += diffs[i].Text
		case Noop:
			// Upon reaching an equality, check for prior redundancies.
			if ndel >= 1 && nins >= 1 {
				// Delete the offending records and add the merged ones.
				i -= ndel + nins
				diffs = splice(diffs, i, ndel+nins)
				a := diffMain(dmp, delStr, insStr, false, deadline)
				for j := len(a) - 1; j >= 0; j-- {
					diffs = splice(diffs, i, 0, a[j])
				}
				i += len(a)
			}

			nins = 0
			ndel = 0
			delStr = ""
			insStr = ""
		}
		i++
	}

	return diffs[:len(diffs)-1] // Remove the dummy entry at the end.
}

// diffBisect finds the 'middle snake' of a diff, splits the problem in two
// and returns the recursively constructed diff.
// See Myers's 1986 paper: An O(ND) Difference Algorithm and Its Variations.
func diffBisect(dmp *DMP, s1, s2 []rune, deadline time.Time) []Diff {
	// Cache the text lengths to prevent multiple calls.
	len1, len2 := len(s1), len(s2)

	dmax := (len1 + len2 + 1) / 2
	offset := dmax
	vlen := 2 * dmax

	v1 := make([]int, vlen)
	v2 := make([]int, vlen)
	for i := range v1 {
		v1[i] = -1
		v2[i] = -1
	}
	v1[offset+1] = 0
	v2[offset+1] = 0

	delta := len1 - len2
	// If the total number of characters is odd, then the front path will
	// collide with the reverse path.
	front := delta%2 != 0
	// Offsets for start and end of k loop.
	// Prevents mapping of space beyond the grid.
	k1start := 0
	k1end := 0
	k2start := 0
	k2end := 0
	for d := 0; d < dmax; d++ {
		// Bail out if deadline is reached.
		if time.Now().After(deadline) {
			break
		}

		// Walk the front path one step.
		for k1 := -d + k1start; k1 <= d-k1end; k1 += 2 {
			k1Offset := offset + k1
			var x1 int

			if k1 == -d || (k1 != d && v1[k1Offset-1] < v1[k1Offset+1]) {
				x1 = v1[k1Offset+1]
			} else {
				x1 = v1[k1Offset-1] + 1
			}

			y1 := x1 - k1
			for x1 < len1 && y1 < len2 {
				if s1[x1] != s2[y1] {
					break
				}
				x1++
				y1++
			}
			v1[k1Offset] = x1
			if x1 > len1 {
				// Ran off the right of the graph.
				k1end += 2
			} else if y1 > len2 {
				// Ran off the bottom of the graph.
				k1start += 2
			} else if front {
				k2Offset := offset + delta - k1
				if k2Offset >= 0 && k2Offset < vlen &&
					v2[k2Offset] != -1 {
					// Mirror x2 onto top-left coordinate system.
					x2 := len1 - v2[k2Offset]
					if x1 >= x2 {
						// Overlap detected.
						return diffBisectSplit(dmp,
							s1, s2, x1, y1, deadline,
						)
					}
				}
			}
		}
		// Walk the reverse path one step.
		for k2 := -d + k2start; k2 <= d-k2end; k2 += 2 {
			k2Offset := offset + k2
			var x2 int
			if k2 == -d || (k2 != d && v2[k2Offset-1] < v2[k2Offset+1]) {
				x2 = v2[k2Offset+1]
			} else {
				x2 = v2[k2Offset-1] + 1
			}
			var y2 = x2 - k2
			for x2 < len1 && y2 < len2 {
				if s1[len1-x2-1] != s2[len2-y2-1] {
					break
				}
				x2++
				y2++
			}
			v2[k2Offset] = x2
			if x2 > len1 {
				// Ran off the left of the graph.
				k2end += 2
			} else if y2 > len2 {
				// Ran off the top of the graph.
				k2start += 2
			} else if !front {
				k1Offset := offset + delta - k2
				if k1Offset >= 0 && k1Offset < vlen &&
					v1[k1Offset] != -1 {
					x1 := v1[k1Offset]
					y1 := offset + x1 - k1Offset
					// Mirror x2 onto top-left coordinate system.
					x2 = len1 - x2
					if x1 >= x2 {
						// Overlap detected.
						return diffBisectSplit(dmp,
							s1, s2, x1, y1, deadline,
						)
					}
				}
			}
		}
	}
	// Diff took too long and hit the deadline or
	// number of diffs equals number of characters, no commonality at all.
	return []Diff{diffDel(string(s1)), diffIns(string(s2))}
}

func diffBisectSplit(dmp *DMP, s1, s2 []rune, x, y int,
	deadline time.Time) []Diff {
	s1a := s1[:x]
	s2a := s2[:y]
	s1b := s1[x:]
	s2b := s2[y:]

	// Compute both diffs serially.
	diffs := diffMainRunes(dmp, s1a, s2a, false, deadline)
	diffsb := diffMainRunes(dmp, s1b, s2b, false, deadline)
	return append(diffs, diffsb...)
}
