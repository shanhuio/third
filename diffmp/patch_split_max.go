package diffmp

func patchSplitMax(ps []Patch, size, margin int) []Patch {
	for x := 0; x < len(ps); x++ {
		cur := ps[x]
		if cur.length1 <= size {
			continue
		}

		// Remove the big old patch.
		ps = append(ps[:x], ps[x+1:]...)
		x--

		start1 := cur.start1
		start2 := cur.start2
		pre := ""
		for len(cur.diffs) != 0 {
			// Create one of several smaller ps.
			p := Patch{}
			empty := true
			p.start1 = start1 - len(pre)
			p.start2 = start2 - len(pre)
			if len(pre) != 0 {
				p.length1 = len(pre)
				p.length2 = len(pre)
				p.diffs = append(p.diffs, Diff{Noop, pre})
			}
			for len(cur.diffs) != 0 && p.length1 < size-margin {
				t := cur.diffs[0].Type
				s := cur.diffs[0].Text
				if t == Insert {
					// Insertions are harmless.
					p.length2 += len(s)
					start2 += len(s)
					p.diffs = append(p.diffs, cur.diffs[0])
					cur.diffs = cur.diffs[1:]
					empty = false
				} else if t == Delete && len(p.diffs) == 1 &&
					p.diffs[0].Type == Noop && len(s) > 2*size {
					// This is a large deletion.
					// Let it pass in one chunk.
					p.length1 += len(s)
					start1 += len(s)
					empty = false
					p.diffs = append(p.diffs, Diff{t, s})
					cur.diffs = cur.diffs[1:]
				} else {
					// Deletion or equality.
					// Only take as much as we can stomach.
					s = s[:min(len(s), size-p.length1-margin)]

					p.length1 += len(s)
					start1 += len(s)
					if t == Noop {
						p.length2 += len(s)
						start2 += len(s)
					} else {
						empty = false
					}
					p.diffs = append(p.diffs, Diff{t, s})
					if s == cur.diffs[0].Text {
						cur.diffs = cur.diffs[1:]
					} else {
						cur.diffs[0].Text = cur.diffs[0].Text[len(s):]
					}
				}
			}
			// Compute the head context for the next patch.
			pre = DiffText2(p.diffs)
			pre = pre[max(0, len(pre)-margin):]

			post := DiffText1(cur.diffs)
			if len(post) > margin {
				post = post[:margin]
			}

			if len(post) != 0 {
				p.length1 += len(post)
				p.length2 += len(post)
				if len(p.diffs) != 0 &&
					p.diffs[len(p.diffs)-1].Type == Noop {
					p.diffs[len(p.diffs)-1].Text += post
				} else {
					p.diffs = append(p.diffs, Diff{Noop, post})
				}
			}
			if !empty {
				x++
				ps = append(ps[:x], append([]Patch{p}, ps[x:]...)...)
			}
		}
	}
	return ps
}
