package gfm

import (
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

var policy = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowDataURIImages()
	return p
}()

// Markdown renders GitHub Flavored Markdown text.
func Markdown(text []byte) []byte {
	unsanitized := blackfriday.MarkdownCommon(text)
	sanitized := policy.SanitizeBytes(unsanitized)
	return sanitized
}
