package diff

import (
	"fmt"
	"strings"
	"testing"
)

func ExampleContextDiffCode() {
	a := "one\ntwo\nthree\nfour\n" + `fmt.Printf("%s,%T",a,b)`
	b := "zero\none\ntree\nfour"
	in := &Input{
		A: &File{
			Lines: SplitLines(a),
			Name:  "Original",
		},
		B: &File{
			Lines: SplitLines(b),
			Name:  "Current",
		},
		Context: 3,
		Eol:     "\n",
	}
	result, _ := ContextDiffString(in)
	fmt.Print(strings.Replace(result, "\t", " ", -1))
	// Output:
	// *** Original
	// --- Current
	// ***************
	// *** 1,5 ****
	//   one
	// ! two
	// ! three
	//   four
	// - fmt.Printf("%s,%T",a,b)
	// --- 1,4 ----
	// + zero
	//   one
	// ! tree
	//   four
}

func ExampleContextDiffString() {
	a := "one\ntwo\nthree\nfour"
	b := "zero\none\ntree\nfour"
	in := &Input{
		A:       NewStringFile("Original", a),
		B:       NewStringFile("Current", b),
		Context: 3,
		Eol:     "\n",
	}
	result, _ := ContextDiffString(in)
	fmt.Printf(strings.Replace(result, "\t", " ", -1))
	// Output:
	// *** Original
	// --- Current
	// ***************
	// *** 1,4 ****
	//   one
	// ! two
	// ! three
	//   four
	// --- 1,4 ----
	// + zero
	//   one
	// ! tree
	//   four
}

func TestOutputFormatRangeFormatContext(t *testing.T) {
	// Per the diff spec at http://www.unix.org/single_unix_specification/
	//
	// The range of lines in file1 shall be written in the following format
	// if the range contains two or more lines:
	//     "*** %d,%d ****\n", <beginning line number>, <ending line number>
	// and the following format otherwise:
	//     "*** %d ****\n", <ending line number>
	// The ending line number of an empty range shall be the number of the preceding line,
	// or 0 if the range is at the start of the file.
	//
	// Next, the range of lines in file2 shall be written in the following format
	// if the range contains two or more lines:
	//     "--- %d,%d ----\n", <beginning line number>, <ending line number>
	// and the following format otherwise:
	//     "--- %d ----\n", <ending line number>
	fm := formatRangeContext
	assertEqual(t, fm(3, 3), "3")
	assertEqual(t, fm(3, 4), "4")
	assertEqual(t, fm(3, 5), "4,5")
	assertEqual(t, fm(3, 6), "4,6")
	assertEqual(t, fm(0, 0), "0")
}
