package diffmp

func splice(s []Diff, off, n int, ds ...Diff) []Diff {
	pre := s[:off]
	post := s[off+n:]

	var ret []Diff
	ret = append(ret, pre...)
	ret = append(ret, ds...)
	ret = append(ret, post...)
	return ret
}
