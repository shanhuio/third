package diff

import (
	"bytes"
	"fmt"
	"io"
)

// WriteContextDiff compares two sequences of lines; generate the delta as a
// context diff.
//
// Context diffs are a compact way of showing line changes and a few lines of
// context. The number of context lines is set by diff.Context which defaults
// to three.
//
// By default, the diff control lines (those with *** or ---) are created with
// a trailing newline.
//
// For inputs that do not have trailing newlines, set the diff.Eol argument to
// "" so that the output will be uniformly newline free.
//
// The context diff format normally has a header for filenames and modification
// times.  Any or all of these may be specified using strings for
// diff.FromFile, diff.ToFile, diff.FromDate, diff.ToDate.  The modification
// times are normally expressed in the ISO 8601 format.  If not specified, the
// strings default to blanks.
func WriteContextDiff(writer io.Writer, in *Input) error {
	var diffErr error
	wf := func(format string, args ...interface{}) {
		_, err := fmt.Fprintf(writer, format, args...)
		if diffErr == nil && err != nil {
			diffErr = err
		}
	}
	ws := func(s string) {
		_, err := fmt.Fprint(writer, s)
		if diffErr == nil && err != nil {
			diffErr = err
		}
	}

	if len(in.Eol) == 0 {
		in.Eol = "\n"
	}

	prefix := map[byte]string{
		'i': "+ ",
		'd': "- ",
		'r': "! ",
		'e': "  ",
	}

	started := false
	m := NewMatcher(in.A, in.B)
	for _, g := range m.GroupedOpCodes(in.Context) {
		if !started {
			started = true
			fromDate := ""
			if len(in.FromDate) > 0 {
				fromDate = "\t" + in.FromDate
			}
			toDate := ""
			if len(in.ToDate) > 0 {
				toDate = "\t" + in.ToDate
			}
			if in.FromFile != "" || in.ToFile != "" {
				wf("*** %s%s%s", in.FromFile, fromDate, in.Eol)
				wf("--- %s%s%s", in.ToFile, toDate, in.Eol)
			}
		}

		first, last := g[0], g[len(g)-1]
		ws("***************" + in.Eol)

		range1 := formatRangeContext(first.I1, last.I2)
		wf("*** %s ****%s", range1, in.Eol)
		for _, c := range g {
			if c.Tag == 'r' || c.Tag == 'd' {
				for _, cc := range g {
					if cc.Tag == 'i' {
						continue
					}
					for _, line := range in.A[cc.I1:cc.I2] {
						ws(prefix[cc.Tag] + line)
					}
				}
				break
			}
		}

		range2 := formatRangeContext(first.J1, last.J2)
		wf("--- %s ----%s", range2, in.Eol)
		for _, c := range g {
			if c.Tag == 'r' || c.Tag == 'i' {
				for _, cc := range g {

					if cc.Tag == 'd' {
						continue
					}
					for _, line := range in.B[cc.J1:cc.J2] {
						ws(prefix[cc.Tag] + line)
					}
				}
				break
			}
		}
	}
	return diffErr
}

// ContextDiffString works like WriteContextDiff but returns the diff a
// string.
func ContextDiffString(in *Input) (string, error) {
	w := new(bytes.Buffer)
	err := WriteContextDiff(w, in)
	return string(w.Bytes()), err
}
