package diff

import (
	"fmt"
	"time"
)

// File is a file for diffing
type File struct {
	Name    string
	Lines   []string
	Time    *time.Time
	TimeStr string
}

// NewStringFile create a file from a string.
func NewStringFile(name, s string) *File {
	return &File{
		Name:  name,
		Lines: SplitLines(s),
	}
}

func (f *File) slice(from, to int) []string {
	return f.Lines[from:to]
}

func (f *File) title() string {
	if f.Time != nil {
		return fmt.Sprintf("%s\t%s", f.Name, f.Time)
	}
	if f.TimeStr != "" {
		return fmt.Sprintf("%s\t%s", f.Name, f.TimeStr)
	}
	return f.Name
}
