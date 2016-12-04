package diffmp

import (
	"strings"
)

func patchAddContext(dmp *DMP, p Patch, s string) Patch {
	if s == "" {
		return p
	}

	pattern := s[p.start2 : p.start2+p.length1]
	padding := 0

	// Look for the first and last matches of pattern in text.  If two
	// different matches are found, increase the pattern length.
	for strings.Index(s, pattern) != strings.LastIndex(s, pattern) &&
		len(pattern) < dmp.MatchMaxBits-2*dmp.PatchMargin {
		padding += dmp.PatchMargin
		maxStart := max(0, p.start2-padding)
		minEnd := min(len(s), p.start2+p.length1+padding)
		pattern = s[maxStart:minEnd]
	}
	// Add one chunk for good luck.
	padding += dmp.PatchMargin

	prefix := s[max(0, p.start2-padding):p.start2]
	suffix := s[p.start2+p.length1 : min(len(s), p.start2+p.length1+padding)]

	if len(prefix) != 0 {
		p.diffs = diffPrepend(diffNoop(prefix), p.diffs)
	}
	if len(suffix) != 0 {
		p.diffs = diffAppend(p.diffs, diffNoop(suffix))
	}

	// Roll back the start points.
	p.start1 -= len(prefix)
	p.start2 -= len(prefix)
	// Extend the lengths.
	p.length1 += len(prefix) + len(suffix)
	p.length2 += len(prefix) + len(suffix)

	return p
}
