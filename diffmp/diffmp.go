package diffmp

import (
	"math"
	"time"
)

// DiffMain finds the differences between two texts.
func (dmp *DMP) DiffMain(s1, s2 string, checkLines bool) []Diff {
	return diffMain(dmp, s1, s2, checkLines, deadline(dmp.DiffTimeout))
}

// DiffMainRunes finds the differences between two rune sequences.
func (dmp *DMP) DiffMainRunes(s1, s2 []rune, checkLines bool) []Diff {
	return diffMainRunes(dmp, s1, s2, checkLines, deadline(dmp.DiffTimeout))
}

// DiffBisect finds the 'middle snake' of a diff, split the problem in two
// and return the recursively constructed diff.
// See Myers 1986 paper: An O(ND) Difference Algorithm and Its Variations.
func (dmp *DMP) DiffBisect(s1, s2 string, deadline time.Time) []Diff {
	// Unused in this code, but retained for interface compatibility.
	return diffBisect(dmp, []rune(s1), []rune(s2), deadline)
}

// DiffHalfMatch checks whether the two texts share a substring which is at
// least half the length of the longer text. This speedup can produce
// non-minimal diffs.
func (dmp *DMP) DiffHalfMatch(s1, s2 string) []string {
	// Unused in this code, but retained for interface compatibility.
	rs := diffHalfMatch(dmp, []rune(s1), []rune(s2))
	if rs == nil {
		return nil
	}

	result := make([]string, len(rs))
	for i, r := range rs {
		result[i] = string(r)
	}
	return result
}

// DiffCleanupEfficiency reduces the number of edits by eliminating
// operationally trivial equalities.
func (dmp *DMP) DiffCleanupEfficiency(diffs []Diff) []Diff {
	return diffCleanupEfficiency(diffs, dmp.DiffEditCost)
}

//  MATCH FUNCTIONS

// MatchMain locates the best instance of 'pattern' in 'text' near 'loc'.
// Returns -1 if no match found.
func (dmp *DMP) MatchMain(s, pattern string, loc int) int {
	// Check for null inputs not needed since null can't be passed in C#.

	loc = int(math.Max(0, math.Min(float64(loc), float64(len(s)))))
	if s == pattern {
		// Shortcut (potentially not guaranteed by the algorithm)
		return 0
	} else if len(s) == 0 {
		// Nothing to match.
		return -1
	} else if loc+len(pattern) <= len(s) &&
		s[loc:loc+len(pattern)] == pattern {
		// Perfect match at the perfect spot!  (Includes case of null pattern)
		return loc
	}
	// Do a fuzzy compare.
	return dmp.MatchBitap(s, pattern, loc)
}

// MatchBitap locates the best instance of 'pattern' in 'text' near 'loc'
// using the Bitap algorithm.  Returns -1 if no match found.
func (dmp *DMP) MatchBitap(text, pattern string, loc int) int {
	return matchBitap(dmp, text, pattern, loc)
}

//  PATCH FUNCTIONS

// PatchAddContext increases the context until it is unique,
// but doesn't let the pattern expand beyond MatchMaxBits.
func (dmp *DMP) PatchAddContext(p Patch, s string) Patch {
	return patchAddContext(dmp, p, s)
}

// PatchMake makes a patch.
func (dmp *DMP) PatchMake(opt ...interface{}) []Patch {
	switch len(opt) {
	case 1:
		diffs, _ := opt[0].([]Diff)
		text1 := DiffText1(diffs)
		return dmp.PatchMake(text1, diffs)

	case 2:
		text1 := opt[0].(string)
		switch t := opt[1].(type) {
		case string:
			diffs := dmp.DiffMain(text1, t, true)
			if len(diffs) > 2 {
				diffs = DiffCleanupSemantic(diffs)
				diffs = dmp.DiffCleanupEfficiency(diffs)
			}
			return dmp.PatchMake(text1, diffs)
		case []Diff:
			return patchMake2(dmp, text1, t)
		}

	case 3:
		return dmp.PatchMake(opt[0], opt[2])
	}
	return []Patch{}
}

// Apply merges a set of patches onto the text.  Returns a patched text,
// as well as an array of true/false values indicating which patches were
// applied.
func (dmp *DMP) Apply(ps []Patch, s string) (string, []bool) {
	if len(ps) == 0 {
		return s, []bool{}
	}

	// Deep copy the patches so that no changes are made to originals.
	ps = PatchDeepCopy(ps)

	nullPadding := patchAddPadding(ps, dmp.PatchMargin)
	s = nullPadding + s + nullPadding
	ps = patchSplitMax(ps, dmp.MatchMaxBits, dmp.PatchMargin)

	x := 0
	// delta keeps track of the offset between the expected and actual
	// location of the previous patch.  If there are patches expected at
	// positions 10 and 20, but the first patch was found at 12, delta is 2
	// and the second patch has an effective expected position of 22.
	delta := 0
	results := make([]bool, len(ps))
	for _, p := range ps {
		expectedLoc := p.start2 + delta
		s1 := DiffText1(p.diffs)
		var startLoc int
		endLoc := -1
		if len(s1) > dmp.MatchMaxBits {
			// PatchSplitMax will only provide an oversized pattern
			// in the case of a monster delete.
			startLoc = dmp.MatchMain(
				s, s1[:dmp.MatchMaxBits], expectedLoc,
			)
			if startLoc != -1 {
				endLoc = dmp.MatchMain(
					s, s1[len(s1)-dmp.MatchMaxBits:],
					expectedLoc+len(s1)-dmp.MatchMaxBits,
				)
				if endLoc == -1 || startLoc >= endLoc {
					// Can't find valid trailing context.  Drop this patch.
					startLoc = -1
				}
			}
		} else {
			startLoc = dmp.MatchMain(s, s1, expectedLoc)
		}
		if startLoc == -1 {
			// No match found.  :(
			results[x] = false
			// Subtract the delta for this failed patch from subsequent
			// patches.
			delta -= p.length2 - p.length1
		} else {
			// Found a match.  :)
			results[x] = true
			delta = startLoc - expectedLoc
			var s2 string
			if endLoc == -1 {
				s2 = s[startLoc:min(startLoc+len(s1), len(s))]
			} else {
				s2 = s[startLoc:min(endLoc+dmp.MatchMaxBits, len(s))]
			}
			if s1 == s2 {
				// Perfect match, just shove the Replacement text in.
				s = s[:startLoc] + DiffText2(p.diffs) +
					s[startLoc+len(s1):]
			} else {
				// Imperfect match.  Run a diff to get a framework of
				// equivalent indices.
				diffs := dmp.DiffMain(s1, s2, false)
				if len(s1) > dmp.MatchMaxBits &&
					float64(DiffLevenshtein(diffs))/float64(len(s1)) >
						dmp.PatchDeleteThreshold {
					// The end points match, but the content is unacceptably
					// bad.
					results[x] = false
				} else {
					diffs = DiffCleanupSemanticLossless(diffs)
					index1 := 0
					for _, d := range p.diffs {
						if d.Type != Noop {
							index2 := DiffXIndex(diffs, index1)
							if d.Type == Insert {
								// Insertion
								s = s[:startLoc+index2] +
									d.Text + s[startLoc+index2:]
							} else if d.Type == Delete {
								// Deletion
								startIndex := startLoc + index2
								s = s[:startIndex] +
									s[startIndex+DiffXIndex(
										diffs,
										index1+len(d.Text),
									)-index2:]
							}
						}
						if d.Type != Delete {
							index1 += len(d.Text)
						}
					}
				}
			}
		}
		x++
	}
	// Strip the padding off.
	s = s[len(nullPadding) : len(nullPadding)+(len(s)-2*len(nullPadding))]
	return s, results
}

// PatchAddPadding adds some padding on text start and end so that edges can
// match something.  Intended to be called only from within patch_apply.
func (dmp *DMP) PatchAddPadding(ps []Patch) string {
	return patchAddPadding(ps, dmp.PatchMargin)
}

// PatchSplitMax looks through the patches and breaks up any which are longer
// than the maximum limit of the match algorithm.
// Intended to be called only from within patch_apply.
func (dmp *DMP) PatchSplitMax(ps []Patch) []Patch {
	return patchSplitMax(ps, dmp.MatchMaxBits, dmp.PatchMargin)
}
