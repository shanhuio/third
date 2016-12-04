package diff

// Input contains the input of two files.
type Input struct {
	A, B    *File
	Eol     string // Headers end of line, defaults to LF
	Context int    // Number of context lines
}
