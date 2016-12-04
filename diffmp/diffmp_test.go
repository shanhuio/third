package diffmp

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchrcom/testify/assert"
)

func Test_runesIndexOf(t *testing.T) {
	target := []rune("abcde")
	for _, test := range []struct {
		pattern string
		start   int
		want    int
	}{
		{"abc", 0, 0},
		{"cde", 0, 2},
		{"e", 0, 4},
		{"cdef", 0, -1},
		{"abcdef", 0, -1},
		{"abc", 2, -1},
		{"cde", 2, 2},
		{"e", 2, 4},
		{"cdef", 2, -1},
		{"abcdef", 2, -1},
		{"e", 6, -1},
	} {
		assert.Equal(t, test.want,
			runesIndexOf(target, []rune(test.pattern), test.start),
			fmt.Sprintf("%q, %d", test.pattern, test.start))
	}
}

func TestDiffHalfmatchTest(t *testing.T) {
	dmp := New()
	dmp.DiffTimeout = 1
	// No match.
	assert.True(t, dmp.DiffHalfMatch("1234567890", "abcdef") == nil, "")
	assert.True(t, dmp.DiffHalfMatch("12345", "23") == nil, "")

	// Single Match.
	assertStrEqual(t,
		[]string{"12", "90", "a", "z", "345678"},
		dmp.DiffHalfMatch("1234567890", "a345678z"))

	assertStrEqual(t, []string{"a", "z", "12", "90", "345678"}, dmp.DiffHalfMatch("a345678z", "1234567890"))

	assertStrEqual(t, []string{"abc", "z", "1234", "0", "56789"}, dmp.DiffHalfMatch("abc56789z", "1234567890"))

	assertStrEqual(t, []string{"a", "xyz", "1", "7890", "23456"}, dmp.DiffHalfMatch("a23456xyz", "1234567890"))

	// Multiple Matches.
	assertStrEqual(t, []string{"12123", "123121", "a", "z", "1234123451234"}, dmp.DiffHalfMatch("121231234123451234123121", "a1234123451234z"))

	assertStrEqual(t, []string{"", "-=-=-=-=-=", "x", "", "x-=-=-=-=-=-=-="}, dmp.DiffHalfMatch("x-=-=-=-=-=-=-=-=-=-=-=-=", "xx-=-=-=-=-=-=-="))

	assertStrEqual(t, []string{"-=-=-=-=-=", "", "", "y", "-=-=-=-=-=-=-=y"}, dmp.DiffHalfMatch("-=-=-=-=-=-=-=-=-=-=-=-=y", "-=-=-=-=-=-=-=yy"))

	// Non-optimal halfmatch.
	// Optimal diff would be -q+x=H-i+e=lloHe+Hu=llo-Hew+y not -qHillo+x=HelloHe-w+Hulloy
	assertStrEqual(t, []string{"qHillo", "w", "x", "Hulloy", "HelloHe"}, dmp.DiffHalfMatch("qHilloHelloHew", "xHelloHeHulloy"))

	// Optimal no halfmatch.
	dmp.DiffTimeout = 0
	assert.True(t, dmp.DiffHalfMatch("qHilloHelloHew", "xHelloHeHulloy") == nil, "")
}

func TestDiffBisectSplit(t *testing.T) {
	// As originally written, this can produce invalid utf8 strings.
	dmp := New()
	diffs := diffBisectSplit(dmp, []rune("STUV\x05WX\x05YZ\x05["),
		[]rune("WĺĻļ\x05YZ\x05ĽľĿŀZ"), 7, 6, time.Now().Add(time.Hour))
	for _, d := range diffs {
		assert.True(t, utf8.ValidString(d.Text))
	}
}

func TestDiffLinesToChars(t *testing.T) {
	// Convert lines down to characters.
	tmpVector := []string{"", "alpha\n", "beta\n"}

	result0, result1, result2 := DiffLinesToChars("alpha\nbeta\nalpha\n", "beta\nalpha\nbeta\n")
	assert.Equal(t, "\u0001\u0002\u0001", result0, "")
	assert.Equal(t, "\u0002\u0001\u0002", result1, "")
	assertStrEqual(t, tmpVector, result2)

	tmpVector = []string{"", "alpha\r\n", "beta\r\n", "\r\n"}
	result0, result1, result2 = DiffLinesToChars("", "alpha\r\nbeta\r\n\r\n\r\n")
	assert.Equal(t, "", result0, "")
	assert.Equal(t, "\u0001\u0002\u0003\u0003", result1, "")
	assertStrEqual(t, tmpVector, result2)

	tmpVector = []string{"", "a", "b"}
	result0, result1, result2 = DiffLinesToChars("a", "b")
	assert.Equal(t, "\u0001", result0, "")
	assert.Equal(t, "\u0002", result1, "")
	assertStrEqual(t, tmpVector, result2)

	// Omit final newline.
	result0, result1, result2 = DiffLinesToChars("alpha\nbeta\nalpha", "")
	assert.Equal(t, "\u0001\u0002\u0003", result0)
	assert.Equal(t, "", result1)
	assertStrEqual(t, []string{"", "alpha\n", "beta\n", "alpha"}, result2)

	// More than 256 to reveal any 8-bit limitations.
	n := 300
	lineList := []string{}
	charList := []rune{}

	for x := 1; x < n+1; x++ {
		lineList = append(lineList, strconv.Itoa(x)+"\n")
		charList = append(charList, rune(x))
	}

	lines := strings.Join(lineList, "")
	chars := string(charList)
	assert.Equal(t, n, utf8.RuneCountInString(chars), "")

	result0, result1, result2 = DiffLinesToChars(lines, "")

	assert.Equal(t, chars, result0)
	assert.Equal(t, "", result1, "")
	// Account for the initial empty element of the lines array.
	assertStrEqual(t, append([]string{""}, lineList...), result2)
}

func TestDiffCharsToLines(t *testing.T) {
	// Convert chars up to lines.
	diffs := []Diff{
		{Noop, "\u0001\u0002\u0001"},
		{Insert, "\u0002\u0001\u0002"}}

	tmpVector := []string{"", "alpha\n", "beta\n"}
	actual := DiffCharsToLines(diffs, tmpVector)
	assertDiffEqual(t, []Diff{
		{Noop, "alpha\nbeta\nalpha\n"},
		{Insert, "beta\nalpha\nbeta\n"}}, actual)

	// More than 256 to reveal any 8-bit limitations.
	n := 300
	lineList := []string{}
	charList := []rune{}

	for x := 1; x <= n; x++ {
		lineList = append(lineList, strconv.Itoa(x)+"\n")
		charList = append(charList, rune(x))
	}

	assert.Equal(t, n, len(charList))

	lineList = append([]string{""}, lineList...)
	diffs = []Diff{{Delete, string(charList)}}
	actual = DiffCharsToLines(diffs, lineList)
	assertDiffEqual(t, []Diff{
		{Delete, strings.Join(lineList, "")}}, actual)
}

func TestDiffCleanupMerge(t *testing.T) {
	// Cleanup a messy diff.
	// Null case.
	diffs := []Diff{}
	diffs = DiffCleanupMerge(diffs)
	assertDiffEqual(t, []Diff{}, diffs)

	// No Diff case.
	diffs = []Diff{{Noop, "a"}, {Delete, "b"}, {Insert, "c"}}
	diffs = DiffCleanupMerge(diffs)
	assertDiffEqual(t, []Diff{{Noop, "a"}, {Delete, "b"}, {Insert, "c"}}, diffs)

	// Merge equalities.
	diffs = []Diff{{Noop, "a"}, {Noop, "b"}, {Noop, "c"}}
	diffs = DiffCleanupMerge(diffs)
	assertDiffEqual(t, []Diff{{Noop, "abc"}}, diffs)

	// Merge deletions.
	diffs = []Diff{{Delete, "a"}, {Delete, "b"}, {Delete, "c"}}
	diffs = DiffCleanupMerge(diffs)
	assertDiffEqual(t, []Diff{{Delete, "abc"}}, diffs)

	// Merge insertions.
	diffs = []Diff{{Insert, "a"}, {Insert, "b"}, {Insert, "c"}}
	diffs = DiffCleanupMerge(diffs)
	assertDiffEqual(t, []Diff{{Insert, "abc"}}, diffs)

	// Merge interweave.
	diffs = []Diff{{Delete, "a"}, {Insert, "b"}, {Delete, "c"}, {Insert, "d"}, {Noop, "e"}, {Noop, "f"}}
	diffs = DiffCleanupMerge(diffs)
	assertDiffEqual(t, []Diff{{Delete, "ac"}, {Insert, "bd"}, {Noop, "ef"}}, diffs)

	// Prefix and suffix detection.
	diffs = []Diff{{Delete, "a"}, {Insert, "abc"}, {Delete, "dc"}}
	diffs = DiffCleanupMerge(diffs)
	assertDiffEqual(t, []Diff{{Noop, "a"}, {Delete, "d"}, {Insert, "b"}, {Noop, "c"}}, diffs)

	// Prefix and suffix detection with equalities.
	diffs = []Diff{{Noop, "x"}, {Delete, "a"}, {Insert, "abc"}, {Delete, "dc"}, {Noop, "y"}}
	diffs = DiffCleanupMerge(diffs)
	assertDiffEqual(t, []Diff{{Noop, "xa"}, {Delete, "d"}, {Insert, "b"}, {Noop, "cy"}}, diffs)

	// Slide edit left.
	diffs = []Diff{{Noop, "a"}, {Insert, "ba"}, {Noop, "c"}}
	diffs = DiffCleanupMerge(diffs)
	assertDiffEqual(t, []Diff{{Insert, "ab"}, {Noop, "ac"}}, diffs)

	// Slide edit right.
	diffs = []Diff{{Noop, "c"}, {Insert, "ab"}, {Noop, "a"}}
	diffs = DiffCleanupMerge(diffs)

	assertDiffEqual(t, []Diff{{Noop, "ca"}, {Insert, "ba"}}, diffs)

	// Slide edit left recursive.
	diffs = []Diff{{Noop, "a"}, {Delete, "b"}, {Noop, "c"}, {Delete, "ac"}, {Noop, "x"}}
	diffs = DiffCleanupMerge(diffs)
	assertDiffEqual(t, []Diff{{Delete, "abc"}, {Noop, "acx"}}, diffs)

	// Slide edit right recursive.
	diffs = []Diff{{Noop, "x"}, {Delete, "ca"}, {Noop, "c"}, {Delete, "b"}, {Noop, "a"}}
	diffs = DiffCleanupMerge(diffs)
	assertDiffEqual(t, []Diff{{Noop, "xca"}, {Delete, "cba"}}, diffs)
}

func TestDiffCleanupSemanticLossless(t *testing.T) {
	// Slide diffs to match logical boundaries.
	// Null case.
	diffs := []Diff{}
	diffs = DiffCleanupSemanticLossless(diffs)
	assertDiffEqual(t, []Diff{}, diffs)

	// Blank lines.
	diffs = []Diff{
		{Noop, "AAA\r\n\r\nBBB"},
		{Insert, "\r\nDDD\r\n\r\nBBB"},
		{Noop, "\r\nEEE"},
	}

	diffs = DiffCleanupSemanticLossless(diffs)

	assertDiffEqual(t, []Diff{
		{Noop, "AAA\r\n\r\n"},
		{Insert, "BBB\r\nDDD\r\n\r\n"},
		{Noop, "BBB\r\nEEE"}}, diffs)

	// Line boundaries.
	diffs = []Diff{
		{Noop, "AAA\r\nBBB"},
		{Insert, " DDD\r\nBBB"},
		{Noop, " EEE"}}

	diffs = DiffCleanupSemanticLossless(diffs)

	assertDiffEqual(t, []Diff{
		{Noop, "AAA\r\n"},
		{Insert, "BBB DDD\r\n"},
		{Noop, "BBB EEE"}}, diffs)

	// Word boundaries.
	diffs = []Diff{
		{Noop, "The c"},
		{Insert, "ow and the c"},
		{Noop, "at."}}

	diffs = DiffCleanupSemanticLossless(diffs)

	assertDiffEqual(t, []Diff{
		{Noop, "The "},
		{Insert, "cow and the "},
		{Noop, "cat."}}, diffs)

	// Alphanumeric boundaries.
	diffs = []Diff{
		{Noop, "The-c"},
		{Insert, "ow-and-the-c"},
		{Noop, "at."}}

	diffs = DiffCleanupSemanticLossless(diffs)

	assertDiffEqual(t, []Diff{
		{Noop, "The-"},
		{Insert, "cow-and-the-"},
		{Noop, "cat."}}, diffs)

	// Hitting the start.
	diffs = []Diff{
		{Noop, "a"},
		{Delete, "a"},
		{Noop, "ax"}}

	diffs = DiffCleanupSemanticLossless(diffs)

	assertDiffEqual(t, []Diff{
		{Delete, "a"},
		{Noop, "aax"}}, diffs)

	// Hitting the end.
	diffs = []Diff{
		{Noop, "xa"},
		{Delete, "a"},
		{Noop, "a"}}

	diffs = DiffCleanupSemanticLossless(diffs)
	assertDiffEqual(t, []Diff{
		{Noop, "xaa"},
		{Delete, "a"}}, diffs)

	// Sentence boundaries.
	diffs = []Diff{
		{Noop, "The xxx. The "},
		{Insert, "zzz. The "},
		{Noop, "yyy."}}

	diffs = DiffCleanupSemanticLossless(diffs)

	assertDiffEqual(t, []Diff{
		{Noop, "The xxx."},
		{Insert, " The zzz."},
		{Noop, " The yyy."}}, diffs)

	// UTF-8 strings.
	diffs = []Diff{
		{Noop, "The ♕. The "},
		{Insert, "♔. The "},
		{Noop, "♖."}}

	diffs = DiffCleanupSemanticLossless(diffs)

	assertDiffEqual(t, []Diff{
		{Noop, "The ♕."},
		{Insert, " The ♔."},
		{Noop, " The ♖."}}, diffs)

	// Rune boundaries.
	diffs = []Diff{
		{Noop, "♕♕"},
		{Insert, "♔♔"},
		{Noop, "♖♖"}}

	diffs = DiffCleanupSemanticLossless(diffs)

	assertDiffEqual(t, []Diff{
		{Noop, "♕♕"},
		{Insert, "♔♔"},
		{Noop, "♖♖"}}, diffs)
}

func TestDiffCleanupSemantic(t *testing.T) {
	// Cleanup semantically trivial equalities.
	// Null case.
	diffs := []Diff{}
	diffs = DiffCleanupSemantic(diffs)
	assertDiffEqual(t, []Diff{}, diffs)

	// No elimination #1.
	diffs = []Diff{
		{Delete, "ab"},
		{Insert, "cd"},
		{Noop, "12"},
		{Delete, "e"}}
	diffs = DiffCleanupSemantic(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "ab"},
		{Insert, "cd"},
		{Noop, "12"},
		{Delete, "e"}}, diffs)

	// No elimination #2.
	diffs = []Diff{
		{Delete, "abc"},
		{Insert, "ABC"},
		{Noop, "1234"},
		{Delete, "wxyz"}}
	diffs = DiffCleanupSemantic(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "abc"},
		{Insert, "ABC"},
		{Noop, "1234"},
		{Delete, "wxyz"}}, diffs)

	// Simple elimination.
	diffs = []Diff{
		{Delete, "a"},
		{Noop, "b"},
		{Delete, "c"}}
	diffs = DiffCleanupSemantic(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "abc"},
		{Insert, "b"}}, diffs)

	// Backpass elimination.
	diffs = []Diff{
		{Delete, "ab"},
		{Noop, "cd"},
		{Delete, "e"},
		{Noop, "f"},
		{Insert, "g"}}
	diffs = DiffCleanupSemantic(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "abcdef"},
		{Insert, "cdfg"}}, diffs)

	// Multiple eliminations.
	diffs = []Diff{
		{Insert, "1"},
		{Noop, "A"},
		{Delete, "B"},
		{Insert, "2"},
		{Noop, "_"},
		{Insert, "1"},
		{Noop, "A"},
		{Delete, "B"},
		{Insert, "2"}}
	diffs = DiffCleanupSemantic(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "AB_AB"},
		{Insert, "1A2_1A2"}}, diffs)

	// Word boundaries.
	diffs = []Diff{
		{Noop, "The c"},
		{Delete, "ow and the c"},
		{Noop, "at."}}
	diffs = DiffCleanupSemantic(diffs)
	assertDiffEqual(t, []Diff{
		{Noop, "The "},
		{Delete, "cow and the "},
		{Noop, "cat."}}, diffs)

	// No overlap elimination.
	diffs = []Diff{
		{Delete, "abcxx"},
		{Insert, "xxdef"}}
	diffs = DiffCleanupSemantic(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "abcxx"},
		{Insert, "xxdef"}}, diffs)

	// Overlap elimination.
	diffs = []Diff{
		{Delete, "abcxxx"},
		{Insert, "xxxdef"}}
	diffs = DiffCleanupSemantic(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "abc"},
		{Noop, "xxx"},
		{Insert, "def"}}, diffs)

	// Reverse overlap elimination.
	diffs = []Diff{
		{Delete, "xxxabc"},
		{Insert, "defxxx"}}
	diffs = DiffCleanupSemantic(diffs)
	assertDiffEqual(t, []Diff{
		{Insert, "def"},
		{Noop, "xxx"},
		{Delete, "abc"}}, diffs)

	// Two overlap eliminations.
	diffs = []Diff{
		{Delete, "abcd1212"},
		{Insert, "1212efghi"},
		{Noop, "----"},
		{Delete, "A3"},
		{Insert, "3BC"}}
	diffs = DiffCleanupSemantic(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "abcd"},
		{Noop, "1212"},
		{Insert, "efghi"},
		{Noop, "----"},
		{Delete, "A"},
		{Noop, "3"},
		{Insert, "BC"}}, diffs)
}

func TestDiffCleanupEfficiency(t *testing.T) {
	dmp := New()
	// Cleanup operationally trivial equalities.
	dmp.DiffEditCost = 4
	// Null case.
	diffs := []Diff{}
	diffs = dmp.DiffCleanupEfficiency(diffs)
	assertDiffEqual(t, []Diff{}, diffs)

	// No elimination.
	diffs = []Diff{
		{Delete, "ab"},
		{Insert, "12"},
		{Noop, "wxyz"},
		{Delete, "cd"},
		{Insert, "34"}}
	diffs = dmp.DiffCleanupEfficiency(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "ab"},
		{Insert, "12"},
		{Noop, "wxyz"},
		{Delete, "cd"},
		{Insert, "34"}}, diffs)

	// Four-edit elimination.
	diffs = []Diff{
		{Delete, "ab"},
		{Insert, "12"},
		{Noop, "xyz"},
		{Delete, "cd"},
		{Insert, "34"}}
	diffs = dmp.DiffCleanupEfficiency(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "abxyzcd"},
		{Insert, "12xyz34"}}, diffs)

	// Three-edit elimination.
	diffs = []Diff{
		{Insert, "12"},
		{Noop, "x"},
		{Delete, "cd"},
		{Insert, "34"}}
	diffs = dmp.DiffCleanupEfficiency(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "xcd"},
		{Insert, "12x34"}}, diffs)

	// Backpass elimination.
	diffs = []Diff{
		{Delete, "ab"},
		{Insert, "12"},
		{Noop, "xy"},
		{Insert, "34"},
		{Noop, "z"},
		{Delete, "cd"},
		{Insert, "56"}}
	diffs = dmp.DiffCleanupEfficiency(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "abxyzcd"},
		{Insert, "12xy34z56"}}, diffs)

	// High cost elimination.
	dmp.DiffEditCost = 5
	diffs = []Diff{
		{Delete, "ab"},
		{Insert, "12"},
		{Noop, "wxyz"},
		{Delete, "cd"},
		{Insert, "34"}}
	diffs = dmp.DiffCleanupEfficiency(diffs)
	assertDiffEqual(t, []Diff{
		{Delete, "abwxyzcd"},
		{Insert, "12wxyz34"}}, diffs)
	dmp.DiffEditCost = 4
}

func TestDiffText(t *testing.T) {
	// Compute the source and destination texts.
	diffs := []Diff{
		{Noop, "jump"},
		{Delete, "s"},
		{Insert, "ed"},
		{Noop, " over "},
		{Delete, "the"},
		{Insert, "a"},
		{Noop, " lazy"}}
	assert.Equal(t, "jumps over the lazy", DiffText1(diffs))
	assert.Equal(t, "jumped over a lazy", DiffText2(diffs))
}

func TestDiffXIndex(t *testing.T) {
	// Translate a location in text1 to text2.
	diffs := []Diff{
		{Delete, "a"},
		{Insert, "1234"},
		{Noop, "xyz"}}
	assert.Equal(t, 5, DiffXIndex(diffs, 2), "diff_xIndex: Translation on equality.")

	diffs = []Diff{
		{Noop, "a"},
		{Delete, "1234"},
		{Noop, "xyz"}}
	assert.Equal(t, 1, DiffXIndex(diffs, 3), "diff_xIndex: Translation on deletion.")
}

func TestDiffLevenshtein(t *testing.T) {
	diffs := []Diff{
		{Delete, "abc"},
		{Insert, "1234"},
		{Noop, "xyz"}}
	assert.Equal(t, 4, DiffLevenshtein(diffs), "diff_levenshtein: Levenshtein with trailing equality.")

	diffs = []Diff{
		{Noop, "xyz"},
		{Delete, "abc"},
		{Insert, "1234"}}
	assert.Equal(t, 4, DiffLevenshtein(diffs), "diff_levenshtein: Levenshtein with leading equality.")

	diffs = []Diff{
		{Delete, "abc"},
		{Noop, "xyz"},
		{Insert, "1234"}}
	assert.Equal(t, 7, DiffLevenshtein(diffs), "diff_levenshtein: Levenshtein with middle equality.")
}

func TestDiffBisect(t *testing.T) {
	dmp := New()
	// Normal.
	a := "cat"
	b := "map"
	// Since the resulting diff hasn't been normalized, it would be ok if
	// the insertion and deletion pairs are swapped.
	// If the order changes, tweak this test as required.
	diffs := []Diff{
		{Delete, "c"},
		{Insert, "m"},
		{Noop, "a"},
		{Delete, "t"},
		{Insert, "p"}}

	assertDiffEqual(t, diffs, dmp.DiffBisect(a, b, time.Date(9999, time.December, 31, 23, 59, 59, 59, time.UTC)))

	// Timeout.
	diffs = []Diff{{Delete, "cat"}, {Insert, "map"}}
	assertDiffEqual(t, diffs, dmp.DiffBisect(a, b, time.Date(0001, time.January, 01, 00, 00, 00, 00, time.UTC)))
}

func TestDiffMain(t *testing.T) {
	dmp := New()
	// Perform a trivial diff.
	diffs := []Diff{}
	assertDiffEqual(t, diffs, dmp.DiffMain("", "", false))

	diffs = []Diff{{Noop, "abc"}}
	assertDiffEqual(t, diffs, dmp.DiffMain("abc", "abc", false))

	diffs = []Diff{{Noop, "ab"}, {Insert, "123"}, {Noop, "c"}}
	assertDiffEqual(t, diffs, dmp.DiffMain("abc", "ab123c", false))

	diffs = []Diff{{Noop, "a"}, {Delete, "123"}, {Noop, "bc"}}
	assertDiffEqual(t, diffs, dmp.DiffMain("a123bc", "abc", false))

	diffs = []Diff{{Noop, "a"}, {Insert, "123"}, {Noop, "b"}, {Insert, "456"}, {Noop, "c"}}
	assertDiffEqual(t, diffs, dmp.DiffMain("abc", "a123b456c", false))

	diffs = []Diff{{Noop, "a"}, {Delete, "123"}, {Noop, "b"}, {Delete, "456"}, {Noop, "c"}}
	assertDiffEqual(t, diffs, dmp.DiffMain("a123b456c", "abc", false))

	// Perform a real diff.
	// Switch off the timeout.
	dmp.DiffTimeout = 0
	diffs = []Diff{{Delete, "a"}, {Insert, "b"}}
	assertDiffEqual(t, diffs, dmp.DiffMain("a", "b", false))

	diffs = []Diff{
		{Delete, "Apple"},
		{Insert, "Banana"},
		{Noop, "s are a"},
		{Insert, "lso"},
		{Noop, " fruit."}}
	assertDiffEqual(t, diffs, dmp.DiffMain("Apples are a fruit.", "Bananas are also fruit.", false))

	diffs = []Diff{
		{Delete, "a"},
		{Insert, "\u0680"},
		{Noop, "x"},
		{Delete, "\t"},
		{Insert, "\u0000"}}
	assertDiffEqual(t, diffs, dmp.DiffMain("ax\t", "\u0680x\u0000", false))

	diffs = []Diff{
		{Delete, "1"},
		{Noop, "a"},
		{Delete, "y"},
		{Noop, "b"},
		{Delete, "2"},
		{Insert, "xab"}}
	assertDiffEqual(t, diffs, dmp.DiffMain("1ayb2", "abxab", false))

	diffs = []Diff{
		{Insert, "xaxcx"},
		{Noop, "abc"}, {Delete, "y"}}
	assertDiffEqual(t, diffs, dmp.DiffMain("abcy", "xaxcxabc", false))

	diffs = []Diff{
		{Delete, "ABCD"},
		{Noop, "a"},
		{Delete, "="},
		{Insert, "-"},
		{Noop, "bcd"},
		{Delete, "="},
		{Insert, "-"},
		{Noop, "efghijklmnopqrs"},
		{Delete, "EFGHIJKLMNOefg"}}
	assertDiffEqual(t, diffs, dmp.DiffMain("ABCDa=bcd=efghijklmnopqrsEFGHIJKLMNOefg", "a-bcd-efghijklmnopqrs", false))

	diffs = []Diff{
		{Insert, " "},
		{Noop, "a"},
		{Insert, "nd"},
		{Noop, " [[Pennsylvania]]"},
		{Delete, " and [[New"}}
	assertDiffEqual(t, diffs, dmp.DiffMain("a [[Pennsylvania]] and [[New", " and [[Pennsylvania]]", false))

	dmp.DiffTimeout = 200 * time.Millisecond // 100ms
	a := "`Twas brillig, and the slithy toves\nDid gyre and gimble in the wabe:\nAll mimsy were the borogoves,\nAnd the mome raths outgrabe.\n"
	b := "I am the very model of a modern major general,\nI've information vegetable, animal, and mineral,\nI know the kings of England, and I quote the fights historical,\nFrom Marathon to Waterloo, in order categorical.\n"
	// Increase the text lengths by 1024 times to ensure a timeout.
	for x := 0; x < 13; x++ {
		a = a + a
		b = b + b
	}

	startTime := time.Now()
	dmp.DiffMain(a, b, true)
	endTime := time.Now()
	delta := endTime.Sub(startTime)
	// Test that we took at least the timeout period.
	assert.True(t, delta >= dmp.DiffTimeout, fmt.Sprintf("%v !>= %v", delta, dmp.DiffTimeout))
	// Test that we didn't take forever (be very forgiving).
	// Theoretically this test could fail very occasionally if the
	// OS task swaps or locks up for a second at the wrong moment.
	// assert.True(t, delta < (dmp.DiffTimeout*3), fmt.Sprintf("%v !< %v", delta, dmp.DiffTimeout*2))
	dmp.DiffTimeout = 0

	// Test the linemode speedup.
	// Must be long to pass the 100 char cutoff.
	a = "1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n"
	b = "abcdefghij\nabcdefghij\nabcdefghij\nabcdefghij\nabcdefghij\nabcdefghij\nabcdefghij\nabcdefghij\nabcdefghij\nabcdefghij\nabcdefghij\nabcdefghij\nabcdefghij\n"
	assertDiffEqual(t, dmp.DiffMain(a, b, true), dmp.DiffMain(a, b, false))

	a = "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
	b = "abcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghijabcdefghij"
	assertDiffEqual(t, dmp.DiffMain(a, b, true), dmp.DiffMain(a, b, false))

	a = "1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n1234567890\n"
	b = "abcdefghij\n1234567890\n1234567890\n1234567890\nabcdefghij\n1234567890\n1234567890\n1234567890\nabcdefghij\n1234567890\n1234567890\n1234567890\nabcdefghij\n"
	textsLinemode := diffRebuildtexts(dmp.DiffMain(a, b, true))
	textsTextmode := diffRebuildtexts(dmp.DiffMain(a, b, false))
	assertStrEqual(t, textsTextmode, textsLinemode)

	// Test null inputs -- not needed because nulls can't be passed in Go.
}

func TestMatchAlphabet(t *testing.T) {
	// Initialise the bitmasks for Bitap.
	bitmask := map[byte]int{
		'a': 4,
		'b': 2,
		'c': 1,
	}
	assertMapEqual(t, bitmask, MatchAlphabet("abc"))

	bitmask = map[byte]int{
		'a': 37,
		'b': 18,
		'c': 8,
	}
	assertMapEqual(t, bitmask, MatchAlphabet("abcaba"))
}

func TestMatchBitap(t *testing.T) {
	dmp := New()

	// Bitap algorithm.
	dmp.MatchDistance = 100
	dmp.MatchThreshold = 0.5
	assert.Equal(t, 5, dmp.MatchBitap("abcdefghijk", "fgh", 5), "match_bitap: Exact match #1.")

	assert.Equal(t, 5, dmp.MatchBitap("abcdefghijk", "fgh", 0), "match_bitap: Exact match #2.")

	assert.Equal(t, 4, dmp.MatchBitap("abcdefghijk", "efxhi", 0), "match_bitap: Fuzzy match #1.")

	assert.Equal(t, 2, dmp.MatchBitap("abcdefghijk", "cdefxyhijk", 5), "match_bitap: Fuzzy match #2.")

	assert.Equal(t, -1, dmp.MatchBitap("abcdefghijk", "bxy", 1), "match_bitap: Fuzzy match #3.")

	assert.Equal(t, 2, dmp.MatchBitap("123456789xx0", "3456789x0", 2), "match_bitap: Overflow.")

	assert.Equal(t, 0, dmp.MatchBitap("abcdef", "xxabc", 4), "match_bitap: Before start match.")

	assert.Equal(t, 3, dmp.MatchBitap("abcdef", "defyy", 4), "match_bitap: Beyond end match.")

	assert.Equal(t, 0, dmp.MatchBitap("abcdef", "xabcdefy", 0), "match_bitap: Oversized pattern.")

	dmp.MatchThreshold = 0.4
	assert.Equal(t, 4, dmp.MatchBitap("abcdefghijk", "efxyhi", 1), "match_bitap: Threshold #1.")

	dmp.MatchThreshold = 0.3
	assert.Equal(t, -1, dmp.MatchBitap("abcdefghijk", "efxyhi", 1), "match_bitap: Threshold #2.")

	dmp.MatchThreshold = 0.0
	assert.Equal(t, 1, dmp.MatchBitap("abcdefghijk", "bcdef", 1), "match_bitap: Threshold #3.")

	dmp.MatchThreshold = 0.5
	assert.Equal(t, 0, dmp.MatchBitap("abcdexyzabcde", "abccde", 3), "match_bitap: Multiple select #1.")

	assert.Equal(t, 8, dmp.MatchBitap("abcdexyzabcde", "abccde", 5), "match_bitap: Multiple select #2.")

	dmp.MatchDistance = 10 // Strict location.
	assert.Equal(t, -1, dmp.MatchBitap("abcdefghijklmnopqrstuvwxyz", "abcdefg", 24), "match_bitap: Distance test #1.")

	assert.Equal(t, 0, dmp.MatchBitap("abcdefghijklmnopqrstuvwxyz", "abcdxxefg", 1), "match_bitap: Distance test #2.")

	dmp.MatchDistance = 1000 // Loose location.
	assert.Equal(t, 0, dmp.MatchBitap("abcdefghijklmnopqrstuvwxyz", "abcdefg", 24), "match_bitap: Distance test #3.")
}

func TestMatchMain(t *testing.T) {
	dmp := New()
	// Full match.
	assert.Equal(t, 0, dmp.MatchMain("abcdef", "abcdef", 1000), "MatchMain: Equality.")

	assert.Equal(t, -1, dmp.MatchMain("", "abcdef", 1), "MatchMain: Null text.")

	assert.Equal(t, 3, dmp.MatchMain("abcdef", "", 3), "MatchMain: Null pattern.")

	assert.Equal(t, 3, dmp.MatchMain("abcdef", "de", 3), "MatchMain: Exact match.")

	assert.Equal(t, 3, dmp.MatchMain("abcdef", "defy", 4), "MatchMain: Beyond end match.")

	assert.Equal(t, 0, dmp.MatchMain("abcdef", "abcdefy", 0), "MatchMain: Oversized pattern.")

	dmp.MatchThreshold = 0.7
	assert.Equal(t, 4, dmp.MatchMain("I am the very model of a modern major general.", " that berry ", 5), "MatchMain: Complex match.")
	dmp.MatchThreshold = 0.5

	// Test null inputs -- not needed because nulls can't be passed in C#.
}

func TestPatchObj(t *testing.T) {
	// Patch Object.
	p := Patch{}
	p.start1 = 20
	p.start2 = 21
	p.length1 = 18
	p.length2 = 17
	p.diffs = []Diff{
		{Noop, "jump"},
		{Delete, "s"},
		{Insert, "ed"},
		{Noop, " over "},
		{Delete, "the"},
		{Insert, "a"},
		{Noop, "\nlaz"}}
	strp := "@@ -21,18 +22,17 @@\n jump\n-s\n+ed\n  over \n-the\n+a\n %0Alaz\n"

	assert.Equal(t, strp, p.String(), "Patch: toString.")
}

func TestPatchFromText(t *testing.T) {
	_v1, _ := PatchFromText("")
	assert.True(t, len(_v1) == 0, "patch_fromText: #0.")
	strp := "@@ -21,18 +22,17 @@\n jump\n-s\n+ed\n  over \n-the\n+a\n %0Alaz\n"
	_v2, _ := PatchFromText(strp)
	assert.Equal(t, strp, _v2[0].String(), "patch_fromText: #1.")

	_v3, _ := PatchFromText("@@ -1 +1 @@\n-a\n+b\n")
	assert.Equal(t, "@@ -1 +1 @@\n-a\n+b\n", _v3[0].String(), "patch_fromText: #2.")

	_v4, _ := PatchFromText("@@ -1,3 +0,0 @@\n-abc\n")
	assert.Equal(t, "@@ -1,3 +0,0 @@\n-abc\n", _v4[0].String(), "patch_fromText: #3.")

	_v5, _ := PatchFromText("@@ -0,0 +1,3 @@\n+abc\n")
	assert.Equal(t, "@@ -0,0 +1,3 @@\n+abc\n", _v5[0].String(), "patch_fromText: #4.")

	// Generates error.
	_, err := PatchFromText("Bad\nPatch\n")
	assert.True(t, err != nil, "There should be an error")
}

func TestPatchToText(t *testing.T) {
	strp := "@@ -21,18 +22,17 @@\n jump\n-s\n+ed\n  over \n-the\n+a\n  laz\n"
	var patches []Patch
	patches, _ = PatchFromText(strp)
	result := PatchToText(patches)
	assert.Equal(t, strp, result)

	strp = "@@ -1,9 +1,9 @@\n-f\n+F\n oo+fooba\n@@ -7,9 +7,9 @@\n obar\n-,\n+.\n  tes\n"
	patches, _ = PatchFromText(strp)
	result = PatchToText(patches)
	assert.Equal(t, strp, result)
}

func TestPatchAddContext(t *testing.T) {
	dmp := New()
	dmp.PatchMargin = 4
	var p Patch
	_p, _ := PatchFromText("@@ -21,4 +21,10 @@\n-jump\n+somersault\n")
	p = _p[0]
	p = dmp.PatchAddContext(p, "The quick brown fox jumps over the lazy dog.")
	assert.Equal(t, "@@ -17,12 +17,18 @@\n fox \n-jump\n+somersault\n s ov\n", p.String(), "patch_addContext: Simple case.")

	_p, _ = PatchFromText("@@ -21,4 +21,10 @@\n-jump\n+somersault\n")
	p = _p[0]
	p = dmp.PatchAddContext(p, "The quick brown fox jumps.")
	assert.Equal(t, "@@ -17,10 +17,16 @@\n fox \n-jump\n+somersault\n s.\n", p.String(), "patch_addContext: Not enough trailing context.")

	_p, _ = PatchFromText("@@ -3 +3,2 @@\n-e\n+at\n")
	p = _p[0]
	p = dmp.PatchAddContext(p, "The quick brown fox jumps.")
	assert.Equal(t, "@@ -1,7 +1,8 @@\n Th\n-e\n+at\n  qui\n", p.String(), "patch_addContext: Not enough leading context.")

	_p, _ = PatchFromText("@@ -3 +3,2 @@\n-e\n+at\n")
	p = _p[0]
	p = dmp.PatchAddContext(p, "The quick brown fox jumps.  The quick brown fox crashes.")
	assert.Equal(t, "@@ -1,27 +1,28 @@\n Th\n-e\n+at\n  quick brown fox jumps. \n", p.String(), "patch_addContext: Ambiguity.")
}

func TestPatchMake(t *testing.T) {
	dmp := New()
	var patches []Patch
	patches = dmp.PatchMake("", "")
	assert.Equal(t, "", PatchToText(patches), "patch_make: Null case.")

	text1 := "The quick brown fox jumps over the lazy dog."
	text2 := "That quick brown fox jumped over a lazy dog."
	expectedPatch := "@@ -1,8 +1,7 @@\n Th\n-at\n+e\n  qui\n@@ -21,17 +21,18 @@\n jump\n-ed\n+s\n  over \n-a\n+the\n  laz\n"
	// The second patch must be "-21,17 +21,18", not "-22,17 +21,18" due to rolling context.
	patches = dmp.PatchMake(text2, text1)
	assert.Equal(t, expectedPatch, PatchToText(patches), "patch_make: Text2+Text1 inputs.")

	expectedPatch = "@@ -1,11 +1,12 @@\n Th\n-e\n+at\n  quick b\n@@ -22,18 +22,17 @@\n jump\n-s\n+ed\n  over \n-the\n+a\n  laz\n"
	patches = dmp.PatchMake(text1, text2)
	assert.Equal(t, expectedPatch, PatchToText(patches), "patch_make: Text1+Text2 inputs.")

	diffs := dmp.DiffMain(text1, text2, false)
	patches = dmp.PatchMake(diffs)
	assert.Equal(t, expectedPatch, PatchToText(patches), "patch_make: Diff input.")

	patches = dmp.PatchMake(text1, diffs)
	assert.Equal(t, expectedPatch, PatchToText(patches), "patch_make: Text1+Diff inputs.")

	patches = dmp.PatchMake(text1, text2, diffs)
	assert.Equal(t, expectedPatch, PatchToText(patches), "patch_make: Text1+Text2+Diff inputs (deprecated).")

	patches = dmp.PatchMake("`1234567890-=[]\\;',./", "~!@#$%^&*()_+{}|:\"<>?")
	assert.Equal(t, "@@ -1,21 +1,21 @@\n-%601234567890-=%5B%5D%5C;',./\n+~!@#$%25%5E&*()_+%7B%7D%7C:%22%3C%3E?\n",
		PatchToText(patches),
		"patch_toText: Character encoding.")

	diffs = []Diff{
		{Delete, "`1234567890-=[]\\;',./"},
		{Insert, "~!@#$%^&*()_+{}|:\"<>?"}}

	_p1, _ := PatchFromText("@@ -1,21 +1,21 @@\n-%601234567890-=%5B%5D%5C;',./\n+~!@#$%25%5E&*()_+%7B%7D%7C:%22%3C%3E?\n")
	assertDiffEqual(t, diffs,
		_p1[0].diffs,
	)

	text1 = ""
	for x := 0; x < 100; x++ {
		text1 += "abcdef"
	}
	text2 = text1 + "123"
	expectedPatch = "@@ -573,28 +573,31 @@\n cdefabcdefabcdefabcdefabcdef\n+123\n"
	patches = dmp.PatchMake(text1, text2)
	assert.Equal(t, expectedPatch, PatchToText(patches), "patch_make: Long string with repeats.")
}

func TestPatchSplitMax(t *testing.T) {
	// Assumes that Match_MaxBits is 32.
	dmp := New()
	var patches []Patch

	patches = dmp.PatchMake("abcdefghijklmnopqrstuvwxyz01234567890", "XabXcdXefXghXijXklXmnXopXqrXstXuvXwxXyzX01X23X45X67X89X0")
	patches = dmp.PatchSplitMax(patches)
	assert.Equal(t, "@@ -1,32 +1,46 @@\n+X\n ab\n+X\n cd\n+X\n ef\n+X\n gh\n+X\n ij\n+X\n kl\n+X\n mn\n+X\n op\n+X\n qr\n+X\n st\n+X\n uv\n+X\n wx\n+X\n yz\n+X\n 012345\n@@ -25,13 +39,18 @@\n zX01\n+X\n 23\n+X\n 45\n+X\n 67\n+X\n 89\n+X\n 0\n", PatchToText(patches))

	patches = dmp.PatchMake("abcdef1234567890123456789012345678901234567890123456789012345678901234567890uvwxyz", "abcdefuvwxyz")
	oldToText := PatchToText(patches)
	dmp.PatchSplitMax(patches)
	assert.Equal(t, oldToText, PatchToText(patches))

	patches = dmp.PatchMake("1234567890123456789012345678901234567890123456789012345678901234567890", "abc")
	patches = dmp.PatchSplitMax(patches)
	assert.Equal(t, "@@ -1,32 +1,4 @@\n-1234567890123456789012345678\n 9012\n@@ -29,32 +1,4 @@\n-9012345678901234567890123456\n 7890\n@@ -57,14 +1,3 @@\n-78901234567890\n+abc\n", PatchToText(patches))

	patches = dmp.PatchMake("abcdefghij , h : 0 , t : 1 abcdefghij , h : 0 , t : 1 abcdefghij , h : 0 , t : 1", "abcdefghij , h : 1 , t : 1 abcdefghij , h : 1 , t : 1 abcdefghij , h : 0 , t : 1")
	dmp.PatchSplitMax(patches)
	assert.Equal(t, "@@ -2,32 +2,32 @@\n bcdefghij , h : \n-0\n+1\n  , t : 1 abcdef\n@@ -29,32 +29,32 @@\n bcdefghij , h : \n-0\n+1\n  , t : 1 abcdef\n", PatchToText(patches))
}

func TestPatchAddPadding(t *testing.T) {
	dmp := New()
	var patches []Patch
	patches = dmp.PatchMake("", "test")
	pass := assert.Equal(t, "@@ -0,0 +1,4 @@\n+test\n",
		PatchToText(patches),
		"PatchAddPadding: Both edges full.")
	if !pass {
		t.FailNow()
	}

	dmp.PatchAddPadding(patches)
	assert.Equal(t, "@@ -1,8 +1,12 @@\n %01%02%03%04\n+test\n %01%02%03%04\n",
		PatchToText(patches),
		"PatchAddPadding: Both edges full.")

	patches = dmp.PatchMake("XY", "XtestY")
	assert.Equal(t, "@@ -1,2 +1,6 @@\n X\n+test\n Y\n",
		PatchToText(patches),
		"PatchAddPadding: Both edges partial.")
	dmp.PatchAddPadding(patches)
	assert.Equal(t, "@@ -2,8 +2,12 @@\n %02%03%04X\n+test\n Y%01%02%03\n",
		PatchToText(patches),
		"PatchAddPadding: Both edges partial.")

	patches = dmp.PatchMake("XXXXYYYY", "XXXXtestYYYY")
	assert.Equal(t, "@@ -1,8 +1,12 @@\n XXXX\n+test\n YYYY\n",
		PatchToText(patches),
		"PatchAddPadding: Both edges none.")
	dmp.PatchAddPadding(patches)
	assert.Equal(t, "@@ -5,8 +5,12 @@\n XXXX\n+test\n YYYY\n",
		PatchToText(patches),
		"PatchAddPadding: Both edges none.")
}

func TestApply(t *testing.T) {
	dmp := New()
	dmp.MatchDistance = 1000
	dmp.MatchThreshold = 0.5
	dmp.PatchDeleteThreshold = 0.5
	patches := []Patch{}
	patches = dmp.PatchMake("", "")
	results0, results1 := dmp.Apply(patches, "Hello world.")
	boolArray := results1
	resultStr := fmt.Sprintf("%v\t%v", results0, len(boolArray))
	pass := assert.Equal(t, "Hello world.\t0", resultStr, "patch_apply: Null case.")
	if !pass {
		t.FailNow()
	}

	patches = dmp.PatchMake("The quick brown fox jumps over the lazy dog.", "That quick brown fox jumped over a lazy dog.")
	results0, results1 = dmp.Apply(patches, "The quick brown fox jumps over the lazy dog.")
	boolArray = results1
	resultStr = results0 + "\t" + strconv.FormatBool(boolArray[0]) + "\t" + strconv.FormatBool(boolArray[1])
	assert.Equal(t, "That quick brown fox jumped over a lazy dog.\ttrue\ttrue", resultStr, "patch_apply: Exact match.")

	results0, results1 = dmp.Apply(patches, "The quick red rabbit jumps over the tired tiger.")
	boolArray = results1
	resultStr = results0 + "\t" + strconv.FormatBool(boolArray[0]) + "\t" + strconv.FormatBool(boolArray[1])
	assert.Equal(t, "That quick red rabbit jumped over a tired tiger.\ttrue\ttrue", resultStr, "patch_apply: Partial match.")

	results0, results1 = dmp.Apply(patches, "I am the very model of a modern major general.")
	boolArray = results1
	resultStr = results0 + "\t" + strconv.FormatBool(boolArray[0]) + "\t" + strconv.FormatBool(boolArray[1])
	assert.Equal(t, "I am the very model of a modern major general.\tfalse\tfalse", resultStr, "patch_apply: Failed match.")

	patches = dmp.PatchMake("x1234567890123456789012345678901234567890123456789012345678901234567890y", "xabcy")
	results0, results1 = dmp.Apply(patches, "x123456789012345678901234567890-----++++++++++-----123456789012345678901234567890y")
	boolArray = results1
	resultStr = results0 + "\t" + strconv.FormatBool(boolArray[0]) + "\t" + strconv.FormatBool(boolArray[1])
	assert.Equal(t, "xabcy\ttrue\ttrue", resultStr, "patch_apply: Big delete, small Diff.")

	patches = dmp.PatchMake("x1234567890123456789012345678901234567890123456789012345678901234567890y", "xabcy")
	results0, results1 = dmp.Apply(patches, "x12345678901234567890---------------++++++++++---------------12345678901234567890y")
	boolArray = results1
	resultStr = results0 + "\t" + strconv.FormatBool(boolArray[0]) + "\t" + strconv.FormatBool(boolArray[1])
	assert.Equal(t, "xabc12345678901234567890---------------++++++++++---------------12345678901234567890y\tfalse\ttrue", resultStr, "patch_apply: Big delete, big Diff 1.")

	dmp.PatchDeleteThreshold = 0.6
	patches = dmp.PatchMake("x1234567890123456789012345678901234567890123456789012345678901234567890y", "xabcy")
	results0, results1 = dmp.Apply(patches, "x12345678901234567890---------------++++++++++---------------12345678901234567890y")
	boolArray = results1
	resultStr = results0 + "\t" + strconv.FormatBool(boolArray[0]) + "\t" + strconv.FormatBool(boolArray[1])
	assert.Equal(t, "xabcy\ttrue\ttrue", resultStr, "patch_apply: Big delete, big Diff 2.")
	dmp.PatchDeleteThreshold = 0.5

	dmp.MatchThreshold = 0.0
	dmp.MatchDistance = 0
	patches = dmp.PatchMake("abcdefghijklmnopqrstuvwxyz--------------------1234567890", "abcXXXXXXXXXXdefghijklmnopqrstuvwxyz--------------------1234567YYYYYYYYYY890")
	results0, results1 = dmp.Apply(patches, "ABCDEFGHIJKLMNOPQRSTUVWXYZ--------------------1234567890")
	boolArray = results1
	resultStr = results0 + "\t" + strconv.FormatBool(boolArray[0]) + "\t" + strconv.FormatBool(boolArray[1])
	assert.Equal(t, "ABCDEFGHIJKLMNOPQRSTUVWXYZ--------------------1234567YYYYYYYYYY890\tfalse\ttrue", resultStr, "patch_apply: Compensate for failed patch.")
	dmp.MatchThreshold = 0.5
	dmp.MatchDistance = 1000

	patches = dmp.PatchMake("", "test")
	patchStr := PatchToText(patches)
	dmp.Apply(patches, "")
	assert.Equal(t, patchStr, PatchToText(patches), "patch_apply: No side effects.")

	patches = dmp.PatchMake("The quick brown fox jumps over the lazy dog.", "Woof")
	patchStr = PatchToText(patches)
	dmp.Apply(patches, "The quick brown fox jumps over the lazy dog.")
	assert.Equal(t, patchStr, PatchToText(patches), "patch_apply: No side effects with major delete.")

	patches = dmp.PatchMake("", "test")
	results0, results1 = dmp.Apply(patches, "")
	boolArray = results1
	resultStr = results0 + "\t" + strconv.FormatBool(boolArray[0])
	assert.Equal(t, "test\ttrue", resultStr, "patch_apply: Edge exact match.")

	patches = dmp.PatchMake("XY", "XtestY")
	results0, results1 = dmp.Apply(patches, "XY")
	boolArray = results1
	resultStr = results0 + "\t" + strconv.FormatBool(boolArray[0])
	assert.Equal(t, "XtestY\ttrue", resultStr, "patch_apply: Near edge exact match.")

	patches = dmp.PatchMake("y", "y123")
	results0, results1 = dmp.Apply(patches, "x")
	boolArray = results1
	resultStr = results0 + "\t" + strconv.FormatBool(boolArray[0])
	assert.Equal(t, "x123\ttrue", resultStr, "patch_apply: Edge partial match.")
}

func TestIndexOf(t *testing.T) {
	type TestCase struct {
		String   string
		Pattern  string
		Position int
		Expected int
	}
	cases := []TestCase{
		{"hi world", "world", -1, 3},
		{"hi world", "world", 0, 3},
		{"hi world", "world", 1, 3},
		{"hi world", "world", 2, 3},
		{"hi world", "world", 3, 3},
		{"hi world", "world", 4, -1},
		{"abbc", "b", -1, 1},
		{"abbc", "b", 0, 1},
		{"abbc", "b", 1, 1},
		{"abbc", "b", 2, 2},
		{"abbc", "b", 3, -1},
		{"abbc", "b", 4, -1},
		// The greek letter beta is the two-byte sequence of "\u03b2".
		{"a\u03b2\u03b2c", "\u03b2", -1, 1},
		{"a\u03b2\u03b2c", "\u03b2", 0, 1},
		{"a\u03b2\u03b2c", "\u03b2", 1, 1},
		{"a\u03b2\u03b2c", "\u03b2", 3, 3},
		{"a\u03b2\u03b2c", "\u03b2", 5, -1},
		{"a\u03b2\u03b2c", "\u03b2", 6, -1},
	}
	for i, c := range cases {
		actual := indexOf(c.String, c.Pattern, c.Position)
		assert.Equal(t, c.Expected, actual, fmt.Sprintf("TestIndex case %d", i))
	}
}

func TestLastIndexOf(t *testing.T) {
	type TestCase struct {
		String   string
		Pattern  string
		Position int
		Expected int
	}
	cases := []TestCase{
		{"hi world", "world", -1, -1},
		{"hi world", "world", 0, -1},
		{"hi world", "world", 1, -1},
		{"hi world", "world", 2, -1},
		{"hi world", "world", 3, -1},
		{"hi world", "world", 4, -1},
		{"hi world", "world", 5, -1},
		{"hi world", "world", 6, -1},
		{"hi world", "world", 7, 3},
		{"hi world", "world", 8, 3},
		{"abbc", "b", -1, -1},
		{"abbc", "b", 0, -1},
		{"abbc", "b", 1, 1},
		{"abbc", "b", 2, 2},
		{"abbc", "b", 3, 2},
		{"abbc", "b", 4, 2},
		// The greek letter beta is the two-byte sequence of "\u03b2".
		{"a\u03b2\u03b2c", "\u03b2", -1, -1},
		{"a\u03b2\u03b2c", "\u03b2", 0, -1},
		{"a\u03b2\u03b2c", "\u03b2", 1, 1},
		{"a\u03b2\u03b2c", "\u03b2", 3, 3},
		{"a\u03b2\u03b2c", "\u03b2", 5, 3},
		{"a\u03b2\u03b2c", "\u03b2", 6, 3},
	}

	for i, c := range cases {
		actual := lastIndexOf(c.String, c.Pattern, c.Position)
		assert.Equal(t, c.Expected, actual,
			fmt.Sprintf("TestLastIndex case %d", i))
	}
}

func Benchmark_DiffMain(bench *testing.B) {
	dmp := New()
	dmp.DiffTimeout = time.Second
	a := "`Twas brillig, and the slithy toves\nDid gyre and gimble in the wabe:\nAll mimsy were the borogoves,\nAnd the mome raths outgrabe.\n"
	b := "I am the very model of a modern major general,\nI've information vegetable, animal, and mineral,\nI know the kings of England, and I quote the fights historical,\nFrom Marathon to Waterloo, in order categorical.\n"
	// Increase the text lengths by 1024 times to ensure a timeout.
	for x := 0; x < 10; x++ {
		a = a + a
		b = b + b
	}
	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		dmp.DiffMain(a, b, true)
	}
}

func Benchmark_DiffCommonPrefix(b *testing.B) {
	a := "ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ"
	for i := 0; i < b.N; i++ {
		CommonPrefixLen(a, a)
	}
}

func Benchmark_DiffCommonSuffix(b *testing.B) {
	a := "ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ"
	for i := 0; i < b.N; i++ {
		CommonSuffixLen(a, a)
	}
}

func Benchmark_DiffMainLarge(b *testing.B) {
	s1 := readFile("speedtest1.txt", b)
	s2 := readFile("speedtest2.txt", b)
	dmp := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dmp.DiffMain(s1, s2, true)
	}
}

func Benchmark_DiffMainLargeLines(b *testing.B) {
	s1 := readFile("speedtest1.txt", b)
	s2 := readFile("speedtest2.txt", b)
	dmp := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		text1, text2, linearray := DiffLinesToRunes(s1, s2)
		diffs := dmp.DiffMainRunes(text1, text2, false)
		diffs = DiffCharsToLines(diffs, linearray)
	}
}

func readFile(filename string, b *testing.B) string {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		b.Fatal(err)
	}
	return string(bytes)
}
