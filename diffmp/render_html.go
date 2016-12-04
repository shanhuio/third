package diffmp

import (
	"bytes"
	"fmt"
	"html"
	"strings"
)

// RenderHTML converts a []Diff into a pretty RenderHTML report.
// It is intended as an example from which to write one's own
// display functions.
func RenderHTML(diffs []Diff) string {
	buf := new(bytes.Buffer)
	for _, d := range diffs {
		s := html.EscapeString(d.Text)
		s = strings.Replace(s, "\n", "&para;<br>", -1)

		switch d.Type {
		case Insert:
			fmt.Fprintf(buf, `<ins style="background:#e6ffe6;">%s</ins>`, s)
		case Delete:
			fmt.Fprintf(buf, `<del style="background:#ffe6e6;">%s</del>`, s)
		case Noop:
			fmt.Fprintf(buf, `<span>%s</span>`, s)
		}
	}
	return buf.String()
}
