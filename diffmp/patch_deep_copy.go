package diffmp

// PatchDeepCopy returns an array that is identical to a
// given an array of patches.
func PatchDeepCopy(patches []Patch) []Patch {
	ret := []Patch{}
	for _, p := range patches {
		cp := p
		cp.diffs = append([]Diff{}, p.diffs...)
		ret = append(ret, cp)
	}
	return ret
}
