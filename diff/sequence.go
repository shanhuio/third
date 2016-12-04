package diff

// SequenceMatcher compares sequence of strings. The basic
// algorithm predates, and is a little fancier than, an algorithm
// published in the late 1980's by Ratcliff and Obershelp under the
// hyperbolic name "gestalt pattern matching".  The basic idea is to find
// the longest contiguous matching subsequence that contains no "junk"
// elements (R-O doesn't address junk).  The same idea is then applied
// recursively to the pieces of the sequences to the left and to the right
// of the matching subsequence.  This does not yield minimal edit
// sequences, but does tend to yield matches that "look right" to people.
//
// SequenceMatcher tries to compute a "human-friendly diff" between two
// sequences.  Unlike e.g. UNIX(tm) diff, the fundamental notion is the
// longest *contiguous* & junk-free matching subsequence.  That's what
// catches peoples' eyes.  The Windows(tm) windiff has another interesting
// notion, pairing up elements that appear uniquely in each sequence.
// That, and the method here, appear to yield more intuitive difference
// reports than does diff.  This method appears to be the least vulnerable
// to synching up on blocks of "junk lines", though (like blank lines in
// ordinary text files, or maybe "<P>" lines in HTML files).  That may be
// because this is the only method of the 3 that has a *concept* of
// "junk" <wink>.
//
// Timing:  Basic R-O is cubic time worst case and quadratic time expected
// case.  SequenceMatcher is quadratic time for the worst case and has
// expected-case behavior dependent in a complicated way on how many
// elements the sequences have in common; best case time is linear.
type SequenceMatcher struct {
	*matcher
}

// NewMatcher creates a new sequence matcher.
func NewMatcher(a, b []string) *SequenceMatcher {
	m := &SequenceMatcher{
		matcher: &matcher{autoJunk: true},
	}
	m.SetSeqs(a, b)
	return m
}

// NewMatcherWithJunk is, well, I don't know what it does exactly...
// TODO(h8liu)
func NewMatcherWithJunk(
	a, b []string, autoJunk bool, isJunk func(string) bool,
) *SequenceMatcher {
	m := &SequenceMatcher{
		matcher: &matcher{isJunk: isJunk, autoJunk: autoJunk},
	}
	m.SetSeqs(a, b)
	return m
}

// SetSeqs sets two sequences to be compared.
func (m *SequenceMatcher) SetSeqs(a, b []string) {
	m.SetSeq1(a)
	m.SetSeq2(b)
}

// SetSeq1 sets the first sequence to be compared. The second sequence to be
// compared is not changed.
//
// SequenceMatcher computes and caches detailed information about the second
// sequence, so if you want to compare one sequence S against many sequences,
// use .SetSeq2(s) once and call .SetSeq1(x) repeatedly for each of the other
// sequences.
//
// See also SetSeqs() and SetSeq2().
func (m *SequenceMatcher) SetSeq1(a []string) {
	if &a == &m.a {
		return
	}
	m.a = a
	m.matchingBlocks = nil
	m.opCodes = nil
}

// SetSeq2 sets the second sequence to be compared. The first sequence to be
// compared is not changed.
func (m *SequenceMatcher) SetSeq2(b []string) {
	if &b == &m.b {
		return
	}
	m.b = b
	m.matchingBlocks = nil
	m.opCodes = nil
	m.fullBCount = nil
	m.chainB()
}

func (m *SequenceMatcher) chainB() {
	// Populate line -> index mapping
	b2j := map[string][]int{}
	for i, s := range m.b {
		indices := b2j[s]
		indices = append(indices, i)
		b2j[s] = indices
	}

	// purge junk elements
	m.bJunk = make(map[string]bool)
	if m.isJunk != nil {
		for s := range b2j {
			if m.isJunk(s) {
				m.bJunk[s] = true
			}
		}
		for s := range m.bJunk {
			delete(b2j, s)
		}
	}

	// purge remaining popular elements
	popular := map[string]struct{}{}
	n := len(m.b)
	if m.autoJunk && n >= 200 {
		ntest := n/100 + 1
		for s, indices := range b2j {
			if len(indices) > ntest {
				popular[s] = struct{}{}
			}
		}
		for s := range popular {
			delete(b2j, s)
		}
	}
	m.bPopular = popular
	m.b2j = b2j
}

// MatchingBlocks returns a list of triples describing matching
// subsequences.
//
// Each triple is of the form (i, j, n), and means that
// a[i:i+n] == b[j:j+n].  The triples are monotonically increasing in
// i and in j. It's also guaranteed that if (i, j, n) and (i', j', n') are
// adjacent triples in the list, and the second is not the last triple in the
// list, then i+n != i' or j+n != j'. IOW, adjacent triples never describe
// adjacent equal blocks.
//
// The last triple is a dummy, (len(a), len(b), 0), and is the only
// triple with n==0.
func (m *SequenceMatcher) MatchingBlocks() []Match {
	if m.matchingBlocks != nil {
		return m.matchingBlocks
	}

	matched := matchBlocks(m.matcher, 0, len(m.a), 0, len(m.b), nil)
	m.matchingBlocks = nonAdjacent(matched, len(m.a), len(m.b))
	return m.matchingBlocks
}

// OpCodes returns the list of 5-tuples describing how to turn a into b.
//
// Each tuple is of the form (tag, i1, i2, j1, j2).  The first tuple has i1 ==
// j1 == 0, and remaining tuples have i1 == the i2 from the tuple preceding it,
// and likewise for j1 == the previous j2.
//
// The tags are characters, with these meanings:
// - 'r' (rep):  a[i1:i2] should be replaced by b[j1:j2]
// - 'd' (del):  a[i1:i2] should be deleted, j1==j2 in this case.
// - 'i' (ins):  b[j1:j2] should be inserted at a[i1:i1], i1==i2 in this case.
// - 'e' (eq):   a[i1:i2] == b[j1:j2]
func (m *SequenceMatcher) OpCodes() []OpCode {
	if m.opCodes != nil {
		return m.opCodes
	}
	i, j := 0, 0
	matching := m.MatchingBlocks()
	opCodes := make([]OpCode, 0, len(matching))
	for _, m := range matching {
		//  invariant:  we've pumped out correct diffs to change
		//  a[:i] into b[:j], and the next matching block is
		//  a[ai:ai+size] == b[bj:bj+size]. So we need to pump
		//  out a diff to change a[i:ai] into b[j:bj], pump out
		//  the matching block, and move (i,j) beyond the match
		ai, bj, size := m.A, m.B, m.Size
		tag := byte(0)
		if i < ai && j < bj {
			tag = 'r'
		} else if i < ai {
			tag = 'd'
		} else if j < bj {
			tag = 'i'
		}
		if tag > 0 {
			opCodes = append(opCodes, OpCode{tag, i, ai, j, bj})
		}
		i, j = ai+size, bj+size
		// the list of matching blocks is terminated by a
		// sentinel with size 0
		if size > 0 {
			opCodes = append(opCodes, OpCode{'e', ai, i, bj, j})
		}
	}
	m.opCodes = opCodes
	return m.opCodes
}

// GroupedOpCodes isolates change clusters by eliminating ranges with no
// changes.
//
// Return a generator of groups with up to n lines of context.
// Each group is in the same format as returned by OpCodes().
func (m *SequenceMatcher) GroupedOpCodes(n int) [][]OpCode {
	if n < 0 {
		n = 3
	}
	codes := m.OpCodes()
	if len(codes) == 0 {
		codes = []OpCode{{'e', 0, 1, 0, 1}}
	}
	// Fixup leading and trailing groups if they show no changes.
	if codes[0].Tag == 'e' {
		c := &codes[0]
		c.I1 = max(c.I1, c.I2-n)
		c.J1 = max(c.J1, c.J2-n)
	}
	if codes[len(codes)-1].Tag == 'e' {
		c := &codes[len(codes)-1]
		c.I2 = min(c.I2, c.I1+n)
		c.J2 = min(c.J2, c.J1+n)
	}
	nn := n + n
	groups := [][]OpCode{}
	group := []OpCode{}
	for _, c := range codes {
		i1, i2, j1, j2 := c.I1, c.I2, c.J1, c.J2
		// End the current group and start a new one whenever
		// there is a large range with no changes.
		if c.Tag == 'e' && i2-i1 > nn {
			group = append(group, OpCode{c.Tag, i1, min(i2, i1+n),
				j1, min(j2, j1+n)})
			groups = append(groups, group)
			group = []OpCode{}
			i1, j1 = max(i1, i2-n), max(j1, j2-n)
		}
		group = append(group, OpCode{c.Tag, i1, i2, j1, j2})
	}
	if len(group) > 0 && !(len(group) == 1 && group[0].Tag == 'e') {
		groups = append(groups, group)
	}
	return groups
}

// Ratio returns a measure of the sequences' similarity (float in [0,1]).
//
// Where T is the total number of elements in both sequences, and
// M is the number of matches, this is 2.0*M / T.
// Note that this is 1 if the sequences are identical, and 0 if
// they have nothing in common.
//
// .Ratio() is expensive to compute if you haven't already computed
// .MatchingBlocks() or .OpCodes(), in which case you may
// want to try .QuickRatio() or .RealQuickRation() first to get an
// upper bound.
func (m *SequenceMatcher) Ratio() float64 {
	matches := 0
	for _, m := range m.MatchingBlocks() {
		matches += m.Size
	}
	return ratio(matches, len(m.a)+len(m.b))
}

// QuickRatio returns an upper bound on ratio() relatively quickly.
//
// This isn't defined beyond that it is an upper bound on .Ratio(), and
// is faster to compute.
func (m *SequenceMatcher) QuickRatio() float64 {
	return quickRatio(m.a, m.b)
}

// RealQuickRatio returns an upper bound on ratio() very quickly.
// This isn't defined beyond that it is an upper bound on .Ratio(), and
// is faster to compute than either .Ratio() or .QuickRatio().
func (m *SequenceMatcher) RealQuickRatio() float64 {
	return realQuickRatio(m.a, m.b)
}
