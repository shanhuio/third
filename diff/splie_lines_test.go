package diff

import (
	"strings"
	"testing"
)

func TestSplitLines(t *testing.T) {
	allTests := []struct {
		input string
		want  []string
	}{
		{"foo", []string{"foo\n"}},
		{"foo\nbar", []string{"foo\n", "bar\n"}},
		{"foo\nbar\n", []string{"foo\n", "bar\n", "\n"}},
	}
	for _, test := range allTests {
		assertEqual(t, SplitLines(test.input), test.want)
	}
}

func benchmarkSplitLines(b *testing.B, count int) {
	str := strings.Repeat("foo\n", count)
	b.ResetTimer()
	n := 0
	for i := 0; i < b.N; i++ {
		n += len(SplitLines(str))
	}
}

func BenchmarkSplitLines100(b *testing.B) {
	benchmarkSplitLines(b, 100)
}

func BenchmarkSplitLines10000(b *testing.B) {
	benchmarkSplitLines(b, 10000)
}
