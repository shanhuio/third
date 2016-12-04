package diff

import (
	"strings"
)

// SplitLines split a string on "\n" while preserving them. The output can be
// used as input for UnifiedDiff and ContextDiff structures.
func SplitLines(s string) []string {
	lines := strings.SplitAfter(s, "\n")
	lines[len(lines)-1] += "\n"
	return lines
}
