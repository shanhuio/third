package diffmp

import (
	"strings"
)

// commonPrefixLen returns the length of the common prefix of two rune
// slices.
func commonPrefixLen(text1, text2 []rune) int {
	short, long := text1, text2
	if len(short) > len(long) {
		short, long = long, short
	}
	for i, r := range short {
		if r != long[i] {
			return i
		}
	}
	return len(short)
}

// commonSuffixLen returns the length of the common suffix of two rune
// slices.
func commonSuffixLen(text1, text2 []rune) int {
	n1 := len(text1)
	n2 := len(text2)
	n := min(n1, n2)
	for i := 0; i < n; i++ {
		if text1[n1-1-i] != text2[n2-1-i] {
			return i
		}
	}
	return n
}

// CommonPrefixLen determines the common prefix length of two strings.
func CommonPrefixLen(s1, s2 string) int {
	return commonPrefixLen([]rune(s1), []rune(s2))
}

// CommonSuffixLen determines the common suffix length of two strings.
func CommonSuffixLen(s1, s2 string) int {
	return commonSuffixLen([]rune(s1), []rune(s2))
}

// CommonOverlap determines if the suffix of one string is the prefix of
// another.
func CommonOverlap(s1, s2 string) int {
	// Cache the text lengths to prevent multiple calls.
	len1 := len(s1)
	len2 := len(s2)
	// Eliminate the null case.
	if len1 == 0 || len2 == 0 {
		return 0
	}
	// Truncate the longer string.
	if len1 > len2 {
		s1 = s1[len1-len2:]
	} else if len1 < len2 {
		s2 = s2[0:len1]
	}
	n := min(len1, len2)
	// Quick check for the worst case.
	if s1 == s2 {
		return n
	}

	// Start by looking for a single character match
	// and increase length until no match is found.
	// Performance analysis: http://neil.fraser.name/news/2010/11/04/
	best := 0
	length := 1
	for {
		pattern := s1[n-length:]
		found := strings.Index(s2, pattern)
		if found == -1 {
			return best
		}
		length += found
		if found == 0 || s1[n-length:] == s2[0:length] {
			best = length
			length++
		}
	}
}
