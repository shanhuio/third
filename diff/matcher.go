package diff

type matcher struct {
	a, b []string

	isJunk   func(string) bool
	autoJunk bool

	matchingBlocks []Match
	opCodes        []OpCode

	// cached stuff for text b
	b2j      map[string][]int
	bJunk    map[string]bool
	bPopular map[string]struct{}

	fullBCount map[string]int // unused
}
