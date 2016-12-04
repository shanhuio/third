package main

import (
	"fmt"
	"log"
	"strings"

	"shanhu.io/third/diff"
)

const f1 = `
xxx
abcdef
`

const f2 = `
xxx
abc
def
`

func main() {
	d := &diff.Input{
		A: &diff.File{
			Lines: diff.SplitLines(strings.TrimSpace(f1)),
			Name:  "a",
		},
		B: &diff.File{
			Lines: diff.SplitLines(strings.TrimSpace(f2)),
			Name:  "b",
		},
		Context: 3,
	}
	text, err := diff.UnifiedDiffString(d)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(text)
}
