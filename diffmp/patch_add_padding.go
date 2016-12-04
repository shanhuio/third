package diffmp

func patchAddPadding(ps []Patch, npad int) string {
	ret := ""
	for x := 1; x <= npad; x++ {
		ret += string(x)
	}

	// Bump all the ps forward.
	for i := range ps {
		ps[i].start1 += npad
		ps[i].start2 += npad
	}

	// Add some padding on start of first diff.
	p := &ps[0]
	if len(p.diffs) == 0 || p.diffs[0].Type != Noop {
		p.diffs = diffPrepend(diffNoop(ret), p.diffs)
		p.start1 -= npad // Should be 0.
		p.start2 -= npad // Should be 0.
		p.length1 += npad
		p.length2 += npad
	} else if npad > len(p.diffs[0].Text) {
		// Grow first equality.
		extraLength := npad - len(p.diffs[0].Text)
		p.diffs[0].Text = ret[len(p.diffs[0].Text):] + p.diffs[0].Text
		p.start1 -= extraLength
		p.start2 -= extraLength
		p.length1 += extraLength
		p.length2 += extraLength
	}

	// Add some padding on end of last diff.
	last := &ps[len(ps)-1]
	if len(last.diffs) == 0 ||
		last.diffs[len(last.diffs)-1].Type != Noop {
		last.diffs = append(last.diffs, Diff{Noop, ret})
		last.length1 += npad
		last.length2 += npad
	} else if npad > len(last.diffs[len(last.diffs)-1].Text) {
		// Grow last equality.
		lastDiff := last.diffs[len(last.diffs)-1]
		extraLength := npad - len(lastDiff.Text)
		last.diffs[len(last.diffs)-1].Text += ret[:extraLength]
		last.length1 += extraLength
		last.length2 += extraLength
	}

	return ret
}
