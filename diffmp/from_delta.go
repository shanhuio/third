package diffmp

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

// FromDelta takes the original s, which is an encoded string that
// describes the operations required to transform text1 into text2,
// and returns the full diff.
func FromDelta(s, delta string) ([]Diff, error) {
	diffs := []Diff{}
	i := 0 // Cursor in text1
	toks := strings.Split(delta, "\t")

	for _, tok := range toks {
		if len(tok) == 0 {
			// Blank tokens are ok (from a trailing \t).
			continue
		}

		// Each token begins with a one character parameter which specifies
		// the operation of this token (delete, insert, equality).
		param := tok[1:]

		switch op := tok[0]; op {
		case '+':
			// decode would Diff all "+" to " "
			param = strings.Replace(param, "+", "%2b", -1)
			var err error
			param, err = url.QueryUnescape(param)
			if err != nil {
				return nil, err
			}
			if !utf8.ValidString(param) {
				return nil, fmt.Errorf("invalid UTF-8 token: %q", param)
			}
			diffs = append(diffs, Diff{Insert, param})
		case '=', '-':
			n, err := strconv.ParseInt(param, 10, 0)
			if err != nil {
				return diffs, err
			} else if n < 0 {
				return diffs, fmt.Errorf(
					"Negative number in DiffFromDelta: %s", param,
				)
			}

			// remember that string slicing is by byte - we want by rune here.
			runes := []rune(s)
			if i+int(n) > len(runes) {
				return diffs, fmt.Errorf("Index out of bound")
			}
			text := string(runes[i : i+int(n)])
			i += int(n)

			if op == '=' {
				diffs = append(diffs, Diff{Noop, text})
			} else {
				diffs = append(diffs, Diff{Delete, text})
			}
		default:
			// Anything else is an error.
			return diffs, fmt.Errorf(
				"Invalid diff operation in DiffFromDelta: %s",
				string(tok[0]),
			)
		}
	}

	if i != len([]rune(s)) {
		return diffs, fmt.Errorf(
			"Delta length (%d) smaller than source text length (%d)",
			i, len(s),
		)
	}
	return diffs, nil
}
