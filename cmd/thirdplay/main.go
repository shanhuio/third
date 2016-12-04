package main

import (
	"fmt"
	"log"

	"shanhu.io/third/diff"
)

const f1 = `
abcdef
`

const f2 = `
abc
def
`

func main() {
	d := &diff.UnifiedDiff{
		A:        diff.SplitLines(f1),
		B:        diff.SplitLines(f2),
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
