package diffmp

import (
	"fmt"
	"testing"

	"github.com/stretchrcom/testify/assert"
)

func TestCommonPrefixLen(t *testing.T) {
	for _, test := range []struct {
		s1, s2 string
		want   int
	}{
		{"abc", "xyz", 0},
		{"1234abcdef", "1234xyz", 4},
		{"1234", "1234xyz", 4},
	} {
		assert.Equal(t, test.want, commonPrefixLen([]rune(test.s1), []rune(test.s2)),
			fmt.Sprintf("%q, %q", test.s1, test.s2))
	}

	assert.Equal(t, 0, CommonPrefixLen("abc", "xyz"), "'abc' and 'xyz' should not be equal")
	assert.Equal(t, 4, CommonPrefixLen("1234abcdef", "1234xyz"), "")
	assert.Equal(t, 4, CommonPrefixLen("1234", "1234xyz"), "")
}

func TestCommonSuffixLen(t *testing.T) {
	for _, test := range []struct {
		s1, s2 string
		want   int
	}{
		{"abc", "xyz", 0},
		{"abcdef1234", "xyz1234", 4},
		{"1234", "xyz1234", 4},
		{"123", "a3", 1},
	} {
		assert.Equal(t, test.want, commonSuffixLen([]rune(test.s1), []rune(test.s2)),
			fmt.Sprintf("%q, %q", test.s1, test.s2))
	}

	assert.Equal(t, 0, CommonSuffixLen("abc", "xyz"), "")
	assert.Equal(t, 4, CommonSuffixLen("abcdef1234", "xyz1234"), "")
	assert.Equal(t, 4, CommonSuffixLen("1234", "xyz1234"), "")
}

func TestCommonOverlap(t *testing.T) {
	assert.Equal(t, 0, CommonOverlap("", "abcd"), "")
	assert.Equal(t, 3, CommonOverlap("abc", "abcd"), "")
	assert.Equal(t, 0, CommonOverlap("123456", "abcd"), "")
	assert.Equal(t, 3, CommonOverlap("123456xxx", "xxxabcd"), "")

	// Unicode.
	// Some overly clever languages (C#) may treat ligatures as equal to their
	// component letters.  E.g. U+FB01 == 'fi'
	assert.Equal(t, 0, CommonOverlap("fi", "\ufb01i"), "")
}

func BenchmarkCommonPrefixLen(b *testing.B) {
	a := "ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ"
	for i := 0; i < b.N; i++ {
		CommonPrefixLen(a, a)
	}
}

func BenchmarkCommonSuffixLen(b *testing.B) {
	a := "ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ"
	for i := 0; i < b.N; i++ {
		CommonSuffixLen(a, a)
	}
}
