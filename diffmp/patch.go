package diffmp

import (
	"bytes"
	"net/url"
	"strconv"
	"strings"
)

// Patch represents one patch operation.
type Patch struct {
	diffs   []Diff
	start1  int
	start2  int
	length1 int
	length2 int
}

// String emulates GNU diff's format.
// Header: @@ -382,8 +481,9 @@
// Indicies are printed as 1-based, not 0-based.
func (p *Patch) String() string {
	var coords1, coords2 string

	if p.length1 == 0 {
		coords1 = strconv.Itoa(p.start1) + ",0"
	} else if p.length1 == 1 {
		coords1 = strconv.Itoa(p.start1 + 1)
	} else {
		coords1 = strconv.Itoa(p.start1+1) + "," + strconv.Itoa(p.length1)
	}

	if p.length2 == 0 {
		coords2 = strconv.Itoa(p.start2) + ",0"
	} else if p.length2 == 1 {
		coords2 = strconv.Itoa(p.start2 + 1)
	} else {
		coords2 = strconv.Itoa(p.start2+1) + "," + strconv.Itoa(p.length2)
	}

	var text bytes.Buffer
	text.WriteString("@@ -" + coords1 + " +" + coords2 + " @@\n")

	// Escape the body of the patch with %xx notation.
	for _, aDiff := range p.diffs {
		switch aDiff.Type {
		case Insert:
			text.WriteString("+")
		case Delete:
			text.WriteString("-")
		case Noop:
			text.WriteString(" ")
		}

		text.WriteString(
			strings.Replace(url.QueryEscape(aDiff.Text), "+", " ", -1),
		)
		text.WriteString("\n")
	}

	return unescaper.Replace(text.String())
}
