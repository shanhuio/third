package diff

import (
	"bytes"
	"fmt"
	"io"
)

// Convert range to the "ed" format
func formatRangeUnified(start, stop int) string {
	// Per the diff spec at http://www.unix.org/single_unix_specification/
	beginning := start + 1 // lines start numbering with one
	length := stop - start
	if length == 1 {
		return fmt.Sprintf("%d", beginning)
	}
	if length == 0 {
		beginning-- // empty ranges begin at line just before the range
	}
	return fmt.Sprintf("%d,%d", beginning, length)
}

// WriteUnifiedDiff compares two sequences of lines; generate the delta as a
// unified diff.
//
// Unified diffs are a compact way of showing line changes and a few lines of
// context. The number of context lines is set by 'n' which defaults to three.
//
// By default, the diff control lines (those with ---, +++, or @@) are created
// with a trailing newline.  This is helpful so that inputs created from
// file.readlines() result in diffs that are suitable for file.writelines()
// since both the inputs and outputs have trailing newlines.
//
// For inputs that do not have trailing newlines, set the lineterm argument to
// "" so that the output will be uniformly newline free.
//
// The unidiff format normally has a header for filenames and modification
// times.  Any or all of these may be specified using strings for 'fromfile',
// 'tofile', 'fromfiledate', and 'tofiledate'.  The modification times are
// normally expressed in the ISO 8601 format.
func WriteUnifiedDiff(writer io.Writer, in *Input) error {
	wf := func(format string, args ...interface{}) error {
		_, err := fmt.Fprintf(writer, format, args...)
		return err
	}
	ws := func(s string) error {
		_, err := fmt.Fprint(writer, s)
		return err
	}

	if in.Eol == "" {
		in.Eol = "\n"
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
				err := wf("--- %s%s%s", in.FromFile, fromDate, in.Eol)
				if err != nil {
					return err
				}
				err = wf("+++ %s%s%s", in.ToFile, toDate, in.Eol)
				if err != nil {
					return err
				}
			}
		}
		first, last := g[0], g[len(g)-1]
		range1 := formatRangeUnified(first.I1, last.I2)
		range2 := formatRangeUnified(first.J1, last.J2)
		if err := wf("@@ -%s +%s @@%s", range1, range2, in.Eol); err != nil {
			return err
		}
		for _, c := range g {
			i1, i2, j1, j2 := c.I1, c.I2, c.J1, c.J2
			if c.Tag == 'e' {
				for _, line := range in.A[i1:i2] {
					if err := ws(" " + line); err != nil {
						return err
					}
				}
				continue
			}
			if c.Tag == 'r' || c.Tag == 'd' {
				for _, line := range in.A[i1:i2] {
					if err := ws("-" + line); err != nil {
						return err
					}
				}
			}
			if c.Tag == 'r' || c.Tag == 'i' {
				for _, line := range in.B[j1:j2] {
					if err := ws("+" + line); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// UnifiedDiffString works like WriteUnifiedDiff but returns the diff a
// string.
func UnifiedDiffString(in *Input) (string, error) {
	w := &bytes.Buffer{}
	err := WriteUnifiedDiff(w, in)
	return string(w.Bytes()), err
}

// formatRangeContext converts range to the "ed" format.
func formatRangeContext(start, stop int) string {
	// Per the diff spec at http://www.unix.org/single_unix_specification/
	beginning := start + 1 // lines start numbering with one
	length := stop - start
	if length == 0 {
		beginning-- // empty ranges begin at line just before the range
	}
	if length <= 1 {
		return fmt.Sprintf("%d", beginning)
	}
	return fmt.Sprintf("%d,%d", beginning, beginning+length-1)
}
