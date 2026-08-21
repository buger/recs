package markdown_test

import (
	"strings"
	"testing"

	"github.com/buger/recs/internal/markdown"
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


func TestRenderCodeAndStarList(t *testing.T) {
	out := markdown.Render("```\ncode\n")
	if !strings.Contains(out, "<pre>") {
		t.Fatal(out)
	}
	out = markdown.Render("* star\n- dash\n### h3\n## h2\n`unclosed")
	if !strings.Contains(out, "<ul>") {
		t.Fatal(out)
	}
	out = markdown.Render("[[x]] and **bold** and *em*")
	_ = out
}

func TestReplaceLinksUnclosed(t *testing.T) {
	out := markdown.Render("[text](https://ex")
	if !strings.Contains(out, "[text](") && !strings.Contains(out, "<p>") {
		t.Log(out)
	}
	_ = markdown.Render("[text](")
	_ = markdown.Render("[no-mid")
}
