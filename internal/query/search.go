package query

import (
	"path/filepath"
	"strings"

	"crm/internal/record"
	"gopkg.in/yaml.v3"
)

// Search scans frontmatter, body, and filename.
// Implements: SYS-REQ-260820-HJPH SW-REQ-260820-X37F
func Search(recs []*record.Record, q string) []*record.Record {
	terms := strings.Fields(strings.ToLower(strings.TrimSpace(q)))
	if len(terms) == 0 {
		return nil
	}
	var out []*record.Record
	for _, rec := range recs {
		blob := searchable(rec)
		ok := true
		for _, t := range terms {
			if !strings.Contains(blob, t) {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, rec)
		}
	}
	return out
}

// Implements: SYS-REQ-260820-HJPH
func searchable(rec *record.Record) string {
	var b strings.Builder
	if rec.Fields != nil {
		if enc, err := yaml.Marshal(rec.Fields); err == nil {
			b.Write(enc)
		}
	}
	b.WriteString("\n")
	b.WriteString(rec.Body)
	b.WriteString("\n")
	b.WriteString(filepath.Base(rec.Path))
	b.WriteString("\n")
	b.WriteString(rec.ID)
	raw := strings.ToLower(b.String())
	var norm strings.Builder
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '\n' {
			norm.WriteRune(r)
		}
	}
	return raw + "\n" + norm.String()
}
