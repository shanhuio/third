package diff

import (
	"fmt"
	"strings"
	"testing"
)

func ExampleUnifiedDiffCode() {
	a := `one
two
three
four
fmt.Printf("%s,%T",a,b)`
	b := `zero
one
three
four`
	in := &Input{
		A: &File{
			Lines:   SplitLines(a),
			Name:    "Original",
			TimeStr: "2005-01-26 23:30:50",
		},
		B: &File{
			Lines:   SplitLines(b),
			Name:    "Current",
			TimeStr: "2010-04-02 10:20:52",
		},
		Context: 3,
	}
	result, _ := UnifiedDiffString(in)
	fmt.Println(strings.Replace(result, "\t", " ", -1))
	// Output:
	// --- Original 2005-01-26 23:30:50
	// +++ Current 2010-04-02 10:20:52
	// @@ -1,5 +1,4 @@
	// +zero
	//  one
	// -two
	//  three
	//  four
	// -fmt.Printf("%s,%T",a,b)
}

func TestOutputFormatRangeFormatUnified(t *testing.T) {
	// Per the diff spec at http://www.unix.org/single_unix_specification/
	//
	// Each <range> field shall be of the form:
	//   %1d", <beginning line number>  if the range contains exactly one line,
	// and:
	//  "%1d,%1d", <beginning line number>, <number of lines> otherwise.
	// If a range is empty, its beginning line number shall be the number of
	// the line just before the range, or 0 if the empty range starts the file.
	fm := formatRangeUnified
	assertEqual(t, fm(3, 3), "3,0")
	assertEqual(t, fm(3, 4), "4")
	assertEqual(t, fm(3, 5), "4,2")
	assertEqual(t, fm(3, 6), "4,3")
	assertEqual(t, fm(0, 0), "0,0")
}
