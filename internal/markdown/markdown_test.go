package markdown_test

import (
	"strings"
	"testing"

	"crm/internal/markdown"
)

// Verifies: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func TestRender(t *testing.T) {
	html := markdown.Render("# Title\n\n## Sub\n\n### H3\n\nA **bold** and *em* and `code` and [x](https://ex)\n\n- one\n- two\n\n```\nraw <b>\n```\n")
	for _, want := range []string{"<h1>", "<h2>", "<h3>", "<strong>", "<em>", "<code>", "<a href=", "<ul>", "<pre>"} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %s in %s", want, html)
		}
	}
	if !strings.Contains(html, "&lt;b&gt;") {
		t.Fatal("escape")
	}
}
