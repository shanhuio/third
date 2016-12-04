package diffmp

// Compute a list of patches to turn text1 into text2.
// text2 is not provided, diffs are the delta between text1 and text2.
func patchMake2(dmp *DMP, text1 string, diffs []Diff) []Patch {
	// Check for null inputs not needed since null can't be passed in C#.
	ps := []Patch{}
	if len(diffs) == 0 {
		return ps // Get rid of the null case.
	}

	p := Patch{}
	nchar1 := 0 // Number of characters into the text1 string.
	nchar2 := 0 // Number of characters into the text2 string.
	// Start with text1 (prepatch) and apply the diffs until we arrive at
	// text2 (post). We recreate the patches one by one to determine
	// context info.
	pre := text1
	post := text1

	for i, d := range diffs {
		if len(p.diffs) == 0 && d.Type != Noop {
			// A new patch starts here.
			p.start1 = nchar1
			p.start2 = nchar2
		}

		switch d.Type {
		case Insert:
			p.diffs = append(p.diffs, d)
			p.length2 += len(d.Text)
			post = post[:nchar2] + d.Text + post[nchar2:]

		case Delete:
			p.length1 += len(d.Text)
			p.diffs = append(p.diffs, d)
			post = post[:nchar2] + post[nchar2+len(d.Text):]

		case Noop:
			if len(d.Text) <= 2*dmp.PatchMargin &&
				len(p.diffs) != 0 && i != len(diffs)-1 {
				// Small equality inside a patch.
				p.diffs = append(p.diffs, d)
				p.length1 += len(d.Text)
				p.length2 += len(d.Text)
			}

			if len(d.Text) >= 2*dmp.PatchMargin {
				// Time for a new patch.
				if len(p.diffs) != 0 {
					p = patchAddContext(dmp, p, pre)
					ps = append(ps, p)
					p = Patch{}
					// Unlike Unidiff, our patch lists have a rolling context.
					// code.google.com/p/google-diff-match-patch/wiki/Unidiff
					// Update prepatch text & pos to reflect the application
					// of the just completed patch.
					pre = post
					nchar1 = nchar2
				}
			}
		}

		// Update the current character count.
		if d.Type != Insert {
			nchar1 += len(d.Text)
		}
		if d.Type != Delete {
			nchar2 += len(d.Text)
		}
	}

	// Pick up the leftover patch if not empty.
	if len(p.diffs) != 0 {
		p = patchAddContext(dmp, p, pre)
		ps = append(ps, p)
	}

	return ps
}
