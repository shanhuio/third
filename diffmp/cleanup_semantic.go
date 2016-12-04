package diffmp

import (
	"unicode/utf8"
)

// DiffCleanupSemanticLossless looks for single edits surrounded on both
// sides by equalities which can be shifted sideways to align the edit to a
// word boundary.
// e.g: The c<ins>at c</ins>ame. -> The <ins>cat </ins>came.
func DiffCleanupSemanticLossless(diffs []Diff) []Diff {
	/**
	 * Given two strings, compute a score representing whether the internal
	 * boundary falls on logical boundaries.
	 * Scores range from 6 (best) to 0 (worst).
	 * Closure, but does not reference any external variables.
	 * @param {string} one First string.
	 * @param {string} two Second string.
	 * @return {number} The score.
	 * @private
	 */
	diffCleanupSemanticScore := func(one, two string) int {
		if len(one) == 0 || len(two) == 0 {
			// Edges are the best.
			return 6
		}

		// Each port of this function behaves slightly differently due to
		// subtle differences in each language's definition of things like
		// 'whitespace'.  Since this function's purpose is largely cosmetic,
		// the choice has been made to use each language's native features
		// rather than force total conformity.
		rune1, _ := utf8.DecodeLastRuneInString(one)
		rune2, _ := utf8.DecodeRuneInString(two)
		char1 := string(rune1)
		char2 := string(rune2)

		nonAlphaNumeric1 := nonAlphaNumericRegex.MatchString(char1)
		nonAlphaNumeric2 := nonAlphaNumericRegex.MatchString(char2)
		whitespace1 := nonAlphaNumeric1 && whitespaceRegex.MatchString(char1)
		whitespace2 := nonAlphaNumeric2 && whitespaceRegex.MatchString(char2)
		lineBreak1 := whitespace1 && linebreakRegex.MatchString(char1)
		lineBreak2 := whitespace2 && linebreakRegex.MatchString(char2)
		blankLine1 := lineBreak1 && blanklineEndRegex.MatchString(one)
		blankLine2 := lineBreak2 && blanklineEndRegex.MatchString(two)

		if blankLine1 || blankLine2 {
			// Five points for blank lines.
			return 5
		} else if lineBreak1 || lineBreak2 {
			// Four points for line breaks.
			return 4
		} else if nonAlphaNumeric1 && !whitespace1 && whitespace2 {
			// Three points for end of sentences.
			return 3
		} else if whitespace1 || whitespace2 {
			// Two points for whitespace.
			return 2
		} else if nonAlphaNumeric1 || nonAlphaNumeric2 {
			// One point for non-alphanumeric.
			return 1
		}
		return 0
	}

	i := 1

	// Intentionally ignore the first and last element (don't need checking).
	for i < len(diffs)-1 {
		if diffs[i-1].Type == Noop &&
			diffs[i+1].Type == Noop {

			// This is a single edit surrounded by equalities.
			equality1 := diffs[i-1].Text
			edit := diffs[i].Text
			equality2 := diffs[i+1].Text

			// First, shift the edit as far left as possible.
			commonOffset := CommonSuffixLen(equality1, edit)
			if commonOffset > 0 {
				commonString := edit[len(edit)-commonOffset:]
				equality1 = equality1[0 : len(equality1)-commonOffset]
				edit = commonString + edit[:len(edit)-commonOffset]
				equality2 = commonString + equality2
			}

			// Second, step character by character right, looking for the best
			// fit.
			bestEquality1 := equality1
			bestEdit := edit
			bestEquality2 := equality2
			bestScore := diffCleanupSemanticScore(equality1, edit) +
				diffCleanupSemanticScore(edit, equality2)

			for len(edit) != 0 && len(equality2) != 0 {
				_, sz := utf8.DecodeRuneInString(edit)
				if len(equality2) < sz || edit[:sz] != equality2[:sz] {
					break
				}
				equality1 += edit[:sz]
				edit = edit[sz:] + equality2[:sz]
				equality2 = equality2[sz:]
				score := diffCleanupSemanticScore(equality1, edit) +
					diffCleanupSemanticScore(edit, equality2)
					// The >= encourages trailing rather than leading
					// whitespace on edits.
				if score >= bestScore {
					bestScore = score
					bestEquality1 = equality1
					bestEdit = edit
					bestEquality2 = equality2
				}
			}

			if diffs[i-1].Text != bestEquality1 {
				// We have an improvement, save it back to the diff.
				if len(bestEquality1) != 0 {
					diffs[i-1].Text = bestEquality1
				} else {
					diffs = splice(diffs, i-1, 1)
					i--
				}

				diffs[i].Text = bestEdit
				if len(bestEquality2) != 0 {
					diffs[i+1].Text = bestEquality2
				} else {
					diffs = append(diffs[:i+1], diffs[i+2:]...)
					i--
				}
			}
		}
		i++
	}

	return diffs
}

// DiffCleanupSemantic reduces the number of edits by eliminating
// semantically trivial equalities.
func DiffCleanupSemantic(diffs []Diff) []Diff {
	changes := false
	equalities := new(Stack) // Stack of indices where equalities are found.

	var lastequality string
	// Always equal to diffs[equalities[equalitiesLength - 1]][1]
	i := 0

	// Number of characters that changed prior to the equality.
	var insLen1, delLen1 int
	// Number of characters that changed after the equality.
	var insLen2, delLen2 int

	for i < len(diffs) {
		if diffs[i].Type == Noop { // Equality found.
			equalities.Push(i)
			insLen1 = insLen2
			delLen1 = delLen2
			insLen2 = 0
			delLen2 = 0
			lastequality = diffs[i].Text
		} else { // An insertion or deletion.
			if diffs[i].Type == Insert {
				insLen2 += len(diffs[i].Text)
			} else {
				delLen2 += len(diffs[i].Text)
			}
			// Eliminate an equality that is smaller or equal to the edits on
			// both sides of it.
			d1 := max(insLen1, delLen1)
			d2 := max(insLen2, delLen2)
			if len(lastequality) > 0 &&
				(len(lastequality) <= d1) &&
				(len(lastequality) <= d2) {
				// Duplicate record.
				insPoint := equalities.Peek().(int)
				diffs = append(
					diffs[:insPoint],
					append(
						[]Diff{{Delete, lastequality}},
						diffs[insPoint:]...,
					)...,
				)

				// Change second copy to insert.
				diffs[insPoint+1].Type = Insert
				// Throw away the equality we just deleted.
				equalities.Pop()

				if equalities.Len() > 0 {
					equalities.Pop()
					i = equalities.Peek().(int)
				} else {
					i = -1
				}

				insLen1 = 0 // Reset the counters.
				delLen1 = 0
				insLen2 = 0
				delLen2 = 0
				lastequality = ""
				changes = true
			}
		}
		i++
	}

	// Normalize the diff.
	if changes {
		diffs = DiffCleanupMerge(diffs)
	}
	diffs = DiffCleanupSemanticLossless(diffs)
	// Find any overlaps between deletions and insertions.
	// e.g: <del>abcxxx</del><ins>xxxdef</ins>
	//   -> <del>abc</del>xxx<ins>def</ins>
	// e.g: <del>xxxabc</del><ins>defxxx</ins>
	//   -> <ins>def</ins>xxx<del>abc</del>
	// Only extract an overlap if it is as big as the edit ahead or behind it.
	i = 1
	for i < len(diffs) {
		if diffs[i-1].Type == Delete &&
			diffs[i].Type == Insert {
			deletion := diffs[i-1].Text
			insertion := diffs[i].Text
			noverlap1 := CommonOverlap(deletion, insertion)
			noverlap2 := CommonOverlap(insertion, deletion)
			if noverlap1 >= noverlap2 {
				if float64(noverlap1) >= float64(len(deletion))/2 ||
					float64(noverlap1) >= float64(len(insertion))/2 {

					// Overlap found.  Insert an equality and trim the
					// surrounding edits.
					diffs = append(
						diffs[:i],
						append(
							[]Diff{
								{Noop, insertion[:noverlap1]},
							},
							diffs[i:]...,
						)...,
					)
					diffs[i-1].Text =
						deletion[0 : len(deletion)-noverlap1]
					diffs[i+1].Text = insertion[noverlap1:]
					i++
				}
			} else {
				if float64(noverlap2) >= float64(len(deletion))/2 ||
					float64(noverlap2) >= float64(len(insertion))/2 {
					// Reverse overlap found.
					// Insert an equality and swap and trim the surrounding
					// edits.
					overlap := Diff{Noop, insertion[noverlap2:]}
					diffs = append(
						diffs[:i],
						append([]Diff{overlap}, diffs[i:]...)...)
					diffs[i-1].Type = Insert
					diffs[i-1].Text =
						insertion[0 : len(insertion)-noverlap2]
					diffs[i+1].Type = Delete
					diffs[i+1].Text = deletion[noverlap2:]
					i++
				}
			}
			i++
		}
		i++
	}

	return diffs
}
