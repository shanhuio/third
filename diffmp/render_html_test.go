package diffmp

import (
	"testing"
)

func TestRenderHtml(t *testing.T) {
	got := RenderHTML([]Diff{
		{Noop, "a\n"},
		{Delete, "<B>b</B>"},
		{Insert, "c&d"},
	})
	want := "<span>a&para;<br></span>" +
		"<del style=\"background:#ffe6e6;\">&lt;B&gt;b&lt;/B&gt;</del>" +
		"<ins style=\"background:#e6ffe6;\">c&amp;d</ins>"
	if got != want {
		t.Errorf("render html, got %q, want %q", got, want)
	}
}
