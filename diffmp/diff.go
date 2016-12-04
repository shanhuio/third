package diffmp

import (
	"fmt"
)

// Diff represents one diff operation
type Diff struct {
	Type Op
	Text string
}

func (d Diff) String() string {
	return fmt.Sprintf("%s %q", opStr(d.Type), d.Text)
}

func diffNoop(s string) Diff { return Diff{Noop, s} }
func diffDel(s string) Diff  { return Diff{Delete, s} }
func diffIns(s string) Diff  { return Diff{Insert, s} }

func diffPrepend(head Diff, diffs []Diff) []Diff {
	ret := make([]Diff, 0, len(diffs)+1)
	ret = append(ret, head)
	ret = append(ret, diffs...)
	return ret
}

func diffAppend(diffs []Diff, tail Diff) []Diff {
	return append(diffs, tail)
}
