package diffmp

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// PatchFromText parses a textual representation of patches and returns a List
// of Patch objects.
func PatchFromText(textline string) ([]Patch, error) {
	patches := []Patch{}
	if len(textline) == 0 {
		return patches, nil
	}
	text := strings.Split(textline, "\n")
	textPointer := 0
	patchHeader := regexp.MustCompile(
		"^@@ -(\\d+),?(\\d*) \\+(\\d+),?(\\d*) @@$",
	)

	var patch Patch
	sign := uint8(0)
	line := ""
	for textPointer < len(text) {

		if !patchHeader.MatchString(text[textPointer]) {
			err := fmt.Errorf("Invalid patch string: %s", text[textPointer])
			return patches, err
		}

		patch = Patch{}
		m := patchHeader.FindStringSubmatch(text[textPointer])

		patch.start1, _ = strconv.Atoi(m[1])
		if len(m[2]) == 0 {
			patch.start1--
			patch.length1 = 1
		} else if m[2] == "0" {
			patch.length1 = 0
		} else {
			patch.start1--
			patch.length1, _ = strconv.Atoi(m[2])
		}

		patch.start2, _ = strconv.Atoi(m[3])

		if len(m[4]) == 0 {
			patch.start2--
			patch.length2 = 1
		} else if m[4] == "0" {
			patch.length2 = 0
		} else {
			patch.start2--
			patch.length2, _ = strconv.Atoi(m[4])
		}
		textPointer++

		for textPointer < len(text) {
			if len(text[textPointer]) > 0 {
				sign = text[textPointer][0]
			} else {
				textPointer++
				continue
			}

			line = text[textPointer][1:]
			line = strings.Replace(line, "+", "%2b", -1)
			line, _ = url.QueryUnescape(line)
			if sign == '-' {
				// Deletion.
				patch.diffs = append(patch.diffs, Diff{Delete, line})
			} else if sign == '+' {
				// Insertion.
				patch.diffs = append(patch.diffs, Diff{Insert, line})
			} else if sign == ' ' {
				// Minor equality.
				patch.diffs = append(patch.diffs, Diff{Noop, line})
			} else if sign == '@' {
				// Start of next patch.
				break
			} else {
				// WTF?
				return patches, fmt.Errorf(
					"Invalid patch mode %q in: %q", sign, line,
				)
			}
			textPointer++
		}

		patches = append(patches, patch)
	}
	return patches, nil
}
