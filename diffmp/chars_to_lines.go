package diffmp

import (
	"strings"
)

// DiffCharsToLines rehydrates the text in a diff from a string of line hashes
// to real lines of text.
func DiffCharsToLines(diffs []Diff, lineArray []string) []Diff {
	hydrated := make([]Diff, 0, len(diffs))
	for _, d := range diffs {
		chars := d.Text
		text := make([]string, len(chars))

		for i, r := range chars {
			text[i] = lineArray[r]
		}

		d.Text = strings.Join(text, "")
		hydrated = append(hydrated, d)
	}
	return hydrated
}
