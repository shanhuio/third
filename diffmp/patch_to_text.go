package diffmp

import (
	"bytes"
)

// PatchToText takes a list of patches and returns a textual representation.
func PatchToText(patches []Patch) string {
	var text bytes.Buffer
	for _, p := range patches {
		text.WriteString(p.String())
	}
	return text.String()
}
