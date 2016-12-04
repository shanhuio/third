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
		A:        diff.SplitLines(strings.TrimSpace(f1)),
		B:        diff.SplitLines(strings.TrimSpace(f2)),
		FromFile: "a",
		ToFile:   "b",
		Context:  3,
	}
	text, err := diff.UnifiedDiffString(d)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(text)
}
