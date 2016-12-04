package diffmp

import (
	"bytes"
)

// DiffText1 computes and returns the source text (all equalities and
// deletions).
func DiffText1(diffs []Diff) string {
	var ret bytes.Buffer
	for _, d := range diffs {
		if d.Type != Insert {
			ret.WriteString(d.Text)
		}
	}
	return ret.String()
}

// DiffText2 computes and returns the destination text (all equalities and
// insertions).
func DiffText2(diffs []Diff) string {
	var ret bytes.Buffer
	for _, d := range diffs {
		if d.Type != Delete {
			ret.WriteString(d.Text)
		}
	}
	return ret.String()
}
