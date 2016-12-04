package diffmp

// MatchAlphabet initialises the alphabet for the Bitap algorithm.
func MatchAlphabet(pattern string) map[byte]int {
	s := map[byte]int{}
	bs := []byte(pattern)
	for _, b := range bs {
		_, ok := s[b]
		if !ok {
			s[b] = 0
		}
	}
	i := 0

	for _, b := range bs {
		value := s[b] | int(uint(1)<<uint((len(pattern)-i-1)))
		s[b] = value
		i++
	}
	return s
}
