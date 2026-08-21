package markdown

import (
	"html"
	"strings"
)

// Render turns a Markdown body into simple HTML.
// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func Render(src string) string {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	lines := strings.Split(src, "\n")
	var b strings.Builder
	inCode := false
	inList := false
	flushList := func() {
		if inList {
			b.WriteString("</ul>\n")
			inList = false
		}
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inCode {
				b.WriteString("</code></pre>\n")
				inCode = false
			} else {
				flushList()
				b.WriteString("<pre><code>")
				inCode = true
			}
			continue
		}
		if inCode {
			b.WriteString(html.EscapeString(line))
			b.WriteByte('\n')
			continue
		}
		trim := strings.TrimSpace(line)
		if trim == "" {
			flushList()
			continue
		}
		if strings.HasPrefix(trim, "- ") || strings.HasPrefix(trim, "* ") {
			if !inList {
				b.WriteString("<ul>\n")
				inList = true
			}
			b.WriteString("<li>")
			b.WriteString(inline(strings.TrimSpace(trim[2:])))
			b.WriteString("</li>\n")
			continue
		}
		flushList()
		if strings.HasPrefix(trim, "### ") {
			b.WriteString("<h3>")
			b.WriteString(inline(trim[4:]))
			b.WriteString("</h3>\n")
			continue
		}
		if strings.HasPrefix(trim, "## ") {
			b.WriteString("<h2>")
			b.WriteString(inline(trim[3:]))
			b.WriteString("</h2>\n")
			continue
		}
		if strings.HasPrefix(trim, "# ") {
			b.WriteString("<h1>")
			b.WriteString(inline(trim[2:]))
			b.WriteString("</h1>\n")
			continue
		}
		b.WriteString("<p>")
		b.WriteString(inline(trim))
		b.WriteString("</p>\n")
	}
	if inCode {
		b.WriteString("</code></pre>\n")
	}
	flushList()
	return b.String()
}

// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func inline(s string) string {
	s = html.EscapeString(s)
	s = replaceDelim(s, "**", "strong")
	s = replaceDelim(s, "*", "em")
	s = replaceDelim(s, "`", "code")
	s = replaceLinks(s)
	return s
}

// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func replaceDelim(s, delim, tag string) string {
	esc := html.EscapeString(delim)
	for {
		i := strings.Index(s, esc)
		if i < 0 {
			return s
		}
		j := strings.Index(s[i+len(esc):], esc)
		if j < 0 {
			return s
		}
		j += i + len(esc)
		inner := s[i+len(esc) : j]
		s = s[:i] + "<" + tag + ">" + inner + "</" + tag + ">" + s[j+len(esc):]
	}
}

// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func replaceLinks(s string) string {
	for {
		open := strings.Index(s, "[")
		if open < 0 {
			return s
		}
		mid := strings.Index(s[open:], "](")
		if mid < 0 {
			return s
		}
		mid += open
		end := strings.Index(s[mid+2:], ")")
		if end < 0 {
			return s
		}
		end += mid + 2
		text := s[open+1 : mid]
		href := s[mid+2 : end]
		s = s[:open] + `<a href="` + href + `">` + text + `</a>` + s[end+1:]
	}
}
