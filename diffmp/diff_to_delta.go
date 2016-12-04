package diffmp

import (
	"bytes"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

// DiffToDelta crushes the diff into an encoded string which describes the
// operations required to transform text1 into text2.
// E.g. =3\t-2\t+ing  -> Keep 3 chars, delete 2 chars, insert 'ing'.
// Operations are tab-separated.  Inserted text is escaped using %xx
// notation.
func DiffToDelta(diffs []Diff) string {
	var buf bytes.Buffer
	for _, d := range diffs {
		switch d.Type {
		case Insert:
			buf.WriteString("+")
			buf.WriteString(
				strings.Replace(url.QueryEscape(d.Text), "+", " ", -1),
			)
			buf.WriteString("\t")
			break
		case Delete:
			buf.WriteString("-")
			buf.WriteString(strconv.Itoa(utf8.RuneCountInString(d.Text)))
			buf.WriteString("\t")
			break
		case Noop:
			buf.WriteString("=")
			buf.WriteString(strconv.Itoa(utf8.RuneCountInString(d.Text)))
			buf.WriteString("\t")
			break
		}
	}
	delta := buf.String()
	if len(delta) != 0 {
		// Strip off trailing tab character.
		delta = delta[0 : utf8.RuneCountInString(delta)-1]
		delta = unescaper.Replace(delta)
	}
	return delta
}
