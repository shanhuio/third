package diff

import (
	"testing"
)

func TestSequenceMatcherRatio(t *testing.T) {
	s := NewMatcher(splitChars("abcd"), splitChars("bcde"))
	assertEqual(t, s.Ratio(), 0.75)
	assertEqual(t, s.QuickRatio(), 0.75)
	assertEqual(t, s.RealQuickRatio(), 1.0)
}
