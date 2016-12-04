package diffmp

// diffLinesToRunesMunge splits a text into an array of strings.  Reduces the
// texts to a []rune where each Unicode character represents one line.
// We use strings instead of []runes as input mainly because you can't use
// []rune as a map key.
func diffLinesToRunesMunge(
	text string, lineArray *[]string, lineHash map[string]int,
) []rune {
	// Walk the text, pulling out a substring for each line.
	// text.split('\n') would would temporarily double our memory footprint.
	// Modifying text would create many large strings to garbage collect.
	lineStart := 0
	lineEnd := -1
	runes := []rune{}

	for lineEnd < len(text)-1 {
		lineEnd = indexOf(text, "\n", lineStart)

		if lineEnd == -1 {
			lineEnd = len(text) - 1
		}

		line := text[lineStart : lineEnd+1]
		lineStart = lineEnd + 1
		lineValue, ok := lineHash[line]

		if ok {
			runes = append(runes, rune(lineValue))
		} else {
			*lineArray = append(*lineArray, line)
			lineHash[line] = len(*lineArray) - 1
			runes = append(runes, rune(len(*lineArray)-1))
		}
	}

	return runes
}

// DiffLinesToRunes splits two texts into a list of runes.  Each rune
// represents one line.
func DiffLinesToRunes(s1, s2 string) ([]rune, []rune, []string) {
	// '\x00' is a valid character, but various debuggers don't like it.
	// So we'll insert a junk entry to avoid generating a null character.
	lineArray := []string{""}    // e.g. lineArray[4] == 'Hello\n'
	lineHash := map[string]int{} // e.g. lineHash['Hello\n'] == 4

	chars1 := diffLinesToRunesMunge(s1, &lineArray, lineHash)
	chars2 := diffLinesToRunesMunge(s2, &lineArray, lineHash)
	return chars1, chars2, lineArray
}

func diffLinesToRunes(s1, s2 []rune) ([]rune, []rune, []string) {
	return DiffLinesToRunes(string(s1), string(s2))
}

// DiffLinesToChars split two texts into a list of strings.  Reduces the texts
// to a string of hashes where each Unicode character represents one line.
// It's slightly faster to call DiffLinesToRunes first, followed by
// DiffMainRunes.
func DiffLinesToChars(s1, s2 string) (string, string, []string) {
	chars1, chars2, lineArray := DiffLinesToRunes(s1, s2)
	return string(chars1), string(chars2), lineArray
}
