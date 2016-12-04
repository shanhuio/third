package main

import (
	"fmt"

	"shanhu.io/third/diffmp"
)

const f1 = `
abcdef
`

const f2 = `
abc
def
`

func main() {
	config := diffmp.New()
	diffs := config.DiffMain(f1, f2, false)
	for _, d := range diffs {
		fmt.Println(d)
	}
}
