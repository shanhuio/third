package diffmp

import (
	"math"
)

// matchBitapScore computes and returns the score for a match with e errors
// and x location.
func matchBitapScore(
	dmp *DMP, e, x, loc int, pattern string,
) float64 {
	accuracy := float64(e) / float64(len(pattern))
	proximity := float64(abs(loc - x))
	if dmp.MatchDistance == 0 {
		// Dodge divide by zero error.
		if proximity == 0 {
			return accuracy
		}
		return 1.0
	}
	return accuracy + (proximity / float64(dmp.MatchDistance))
}

func matchBitap(dmp *DMP, text, pattern string, loc int) int {
	// Initialise the alphabet.
	s := MatchAlphabet(pattern)

	// Highest score beyond which we give up.
	threshold := float64(dmp.MatchThreshold)
	// Is there a nearby exact match? (speedup)
	bestLoc := indexOf(text, pattern, loc)
	if bestLoc != -1 {
		threshold = math.Min(
			matchBitapScore(dmp, 0, bestLoc, loc, pattern),
			threshold,
		)
		// What about in the other direction? (speedup)
		bestLoc = lastIndexOf(text, pattern, loc+len(pattern))
		if bestLoc != -1 {
			threshold = math.Min(
				matchBitapScore(dmp, 0, bestLoc, loc, pattern),
				threshold,
			)
		}
	}

	// Initialise the bit arrays.
	matchmask := 1 << uint((len(pattern) - 1))
	bestLoc = -1

	var binMin, binMid int
	binMax := len(pattern) + len(text)
	lastRD := []int{}
	for d := 0; d < len(pattern); d++ {
		// Scan for the best match; each iteration allows for one more error.
		// Run a binary search to determine how far from 'loc' we can stray at
		// this error level.
		binMin = 0
		binMid = binMax
		for binMin < binMid {
			if matchBitapScore(
				dmp, d, loc+binMid, loc, pattern,
			) <= threshold {
				binMin = binMid
			} else {
				binMax = binMid
			}
			binMid = (binMax-binMin)/2 + binMin
		}
		// Use the result from this iteration as the maximum for the next.
		binMax = binMid
		start := max(1, loc-binMid+1)
		finish := min(loc+binMid, len(text)) + len(pattern)

		rd := make([]int, finish+2)
		rd[finish+1] = (1 << uint(d)) - 1

		for j := finish; j >= start; j-- {
			var charMatch int
			if len(text) <= j-1 {
				// Out of range.
				charMatch = 0
			} else if _, ok := s[text[j-1]]; !ok {
				charMatch = 0
			} else {
				charMatch = s[text[j-1]]
			}

			if d == 0 {
				// First pass: exact match.
				rd[j] = ((rd[j+1] << 1) | 1) & charMatch
			} else {
				// Subsequent passes: fuzzy match.
				rd[j] = ((rd[j+1]<<1)|1)&charMatch |
					(((lastRD[j+1] | lastRD[j]) << 1) | 1) | lastRD[j+1]
			}
			if (rd[j] & matchmask) != 0 {
				score := matchBitapScore(dmp, d, j-1, loc, pattern)
				// This match will almost certainly be better than any
				// existing match.  But check anyway.
				if score <= threshold {
					// Told you so.
					threshold = score
					bestLoc = j - 1
					if bestLoc > loc {
						// When passing loc, don't exceed our current distance
						// from loc.
						start = max(1, 2*loc-bestLoc)
					} else {
						// Already passed loc, downhill from here on in.
						break
					}
				}
			}
		}
		if matchBitapScore(dmp, d+1, loc, loc, pattern) > threshold {
			// No hope for a (better) match at greater error levels.
			break
		}
		lastRD = rd
	}
	return bestLoc
}
