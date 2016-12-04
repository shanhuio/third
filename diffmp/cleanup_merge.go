package diffmp

import (
	"strings"
)

// DiffCleanupMerge reorders and merges like edit sections.  Merge
// equalities.  Any edit section can move as long as it doesn't cross an
// equality.
func DiffCleanupMerge(ds []Diff) []Diff {
	// Add a dummy entry at the end.
	ds = append(ds, Diff{Noop, ""})
	i := 0
	ndel := 0
	nins := 0
	commonlength := 0
	delStr := ""
	insStr := ""

	for i < len(ds) {
		switch ds[i].Type {
		case Insert:
			nins++
			insStr += ds[i].Text
			i++
			break
		case Delete:
			ndel++
			delStr += ds[i].Text
			i++
			break
		case Noop:
			// Upon reaching an equality, check for prior redundancies.
			if ndel+nins > 1 {
				if ndel != 0 && nins != 0 {
					// Factor out any common prefixies.
					commonlength = CommonPrefixLen(
						insStr, delStr,
					)
					if commonlength != 0 {
						x := i - ndel - nins
						if x > 0 && ds[x-1].Type == Noop {
							ds[x-1].Text += insStr[:commonlength]
						} else {
							ds = append(
								[]Diff{
									{Noop,
										insStr[:commonlength]},
								},
								ds...,
							)
							i++
						}
						insStr = insStr[commonlength:]
						delStr = delStr[commonlength:]
					}
					// Factor out any common suffixies.
					commonlength = CommonSuffixLen(
						insStr, delStr,
					)
					if commonlength != 0 {
						insertIndex := len(insStr) - commonlength
						deleteIndex := len(delStr) - commonlength
						ds[i].Text = insStr[insertIndex:] + ds[i].Text
						insStr = insStr[:insertIndex]
						delStr = delStr[:deleteIndex]
					}
				}
				// Delete the offending records and add the merged ones.
				if ndel == 0 {
					ds = splice(ds, i-nins, ndel+nins, diffIns(insStr))
				} else if nins == 0 {
					ds = splice(ds, i-ndel, ndel+nins, diffDel(delStr))
				} else {
					ds = splice(
						ds, i-ndel-nins, ndel+nins,
						diffDel(delStr), diffIns(insStr),
					)
				}

				i = i - ndel - nins + 1
				if ndel != 0 {
					i++
				}
				if nins != 0 {
					i++
				}
			} else if i != 0 && ds[i-1].Type == Noop {
				// Merge this equality with the previous one.
				ds[i-1].Text += ds[i].Text
				ds = append(ds[:i], ds[i+1:]...)
			} else {
				i++
			}
			nins = 0
			ndel = 0
			delStr = ""
			insStr = ""
			break
		}
	}

	if len(ds[len(ds)-1].Text) == 0 {
		ds = ds[0 : len(ds)-1] // Remove the dummy entry at the end.
	}

	// Second pass: look for single edits surrounded on both sides by
	// equalities which can be shifted sideways to eliminate an equality.
	// e.g: A<ins>BA</ins>C -> <ins>AB</ins>AC
	changes := false
	i = 1
	// Intentionally ignore the first and last element (don't need checking).
	for i < (len(ds) - 1) {
		if ds[i-1].Type == Noop &&
			ds[i+1].Type == Noop {
			// This is a single edit surrounded by equalities.
			if strings.HasSuffix(ds[i].Text, ds[i-1].Text) {
				// Shift the edit over the previous equality.
				ds[i].Text = ds[i-1].Text +
					ds[i].Text[:len(ds[i].Text)-len(ds[i-1].Text)]
				ds[i+1].Text =
					ds[i-1].Text + ds[i+1].Text
				ds = splice(ds, i-1, 1)
				changes = true
			} else if strings.HasPrefix(
				ds[i].Text, ds[i+1].Text,
			) {
				// Shift the edit over the next equality.
				ds[i-1].Text += ds[i+1].Text
				ds[i].Text =
					ds[i].Text[len(ds[i+1].Text):] +
						ds[i+1].Text
				ds = splice(ds, i+1, 1)
				changes = true
			}
		}
		i++
	}

	// If shifts were made, the diff needs reordering and another shift sweep.
	if changes {
		ds = DiffCleanupMerge(ds)
	}

	return ds
}
