package diff

import (
	"bytes"
	"fmt"
	"io"
)

// ContextDiff contains the same input as the unifited diff
// TODO(h8liu): what the fuck?!
type ContextDiff UnifiedDiff

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
func WriteContextDiff(writer io.Writer, diff *ContextDiff) error {
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

	if len(diff.Eol) == 0 {
		diff.Eol = "\n"
	}

	prefix := map[byte]string{
		'i': "+ ",
		'd': "- ",
		'r': "! ",
		'e': "  ",
	}

	started := false
	m := NewMatcher(diff.A, diff.B)
	for _, g := range m.GroupedOpCodes(diff.Context) {
		if !started {
			started = true
			fromDate := ""
			if len(diff.FromDate) > 0 {
				fromDate = "\t" + diff.FromDate
			}
			toDate := ""
			if len(diff.ToDate) > 0 {
				toDate = "\t" + diff.ToDate
			}
			if diff.FromFile != "" || diff.ToFile != "" {
				wf("*** %s%s%s", diff.FromFile, fromDate, diff.Eol)
				wf("--- %s%s%s", diff.ToFile, toDate, diff.Eol)
			}
		}

		first, last := g[0], g[len(g)-1]
		ws("***************" + diff.Eol)

		range1 := formatRangeContext(first.I1, last.I2)
		wf("*** %s ****%s", range1, diff.Eol)
		for _, c := range g {
			if c.Tag == 'r' || c.Tag == 'd' {
				for _, cc := range g {
					if cc.Tag == 'i' {
						continue
					}
					for _, line := range diff.A[cc.I1:cc.I2] {
						ws(prefix[cc.Tag] + line)
					}
				}
				break
			}
		}

		range2 := formatRangeContext(first.J1, last.J2)
		wf("--- %s ----%s", range2, diff.Eol)
		for _, c := range g {
			if c.Tag == 'r' || c.Tag == 'i' {
				for _, cc := range g {

					if cc.Tag == 'd' {
						continue
					}
					for _, line := range diff.B[cc.J1:cc.J2] {
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
func ContextDiffString(diff *ContextDiff) (string, error) {
	w := &bytes.Buffer{}
	err := WriteContextDiff(w, diff)
	return string(w.Bytes()), err
}
