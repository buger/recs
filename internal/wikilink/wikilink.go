package wikilink

import (
	"regexp"
	"strings"

	"crm/internal/record"
)

var wikiRe = regexp.MustCompile(`\[\[([^\]|]+)(?:\|([^\]]+))?\]\]`)

// Match finds a record by display name, title, name, or id.
// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func Match(query string, recs []*record.Record) *record.Record {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil
	}
	ql := strings.ToLower(q)
	var fuzzy *record.Record
	for _, rec := range recs {
		if rec == nil {
			continue
		}
		if rec.ID == q || strings.EqualFold(rec.ID, q) {
			return rec
		}
		name := record.DisplayName(rec)
		if strings.EqualFold(name, q) {
			return rec
		}
		if fuzzy == nil && (strings.Contains(strings.ToLower(name), ql) || strings.Contains(strings.ToLower(rec.ID), ql)) {
			fuzzy = rec
		}
	}
	return fuzzy
}

// Resolve replaces [[Name]] tokens with HTML anchors toward record ids.
// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func Resolve(src string, recs []*record.Record) string {
	return wikiRe.ReplaceAllStringFunc(src, func(m string) string {
		parts := wikiRe.FindStringSubmatch(m)
		if len(parts) < 2 {
			return m
		}
		label := parts[1]
		if parts[2] != "" {
			label = parts[2]
		}
		target := parts[1]
		if rec := Match(target, recs); rec != nil {
			return `<a class="wikilink" href="#/r/` + rec.ID + `">` + label + `</a>`
		}
		return `<span class="wikilink missing">` + label + `</span>`
	})
}

// FindAll returns wikilink target strings in src.
// Implements: SYS-REQ-260821-QF1J
func FindAll(src string) []string {
	matches := wikiRe.FindAllStringSubmatch(src, -1)
	var out []string
	for _, m := range matches {
		if len(m) > 1 {
			out = append(out, m[1])
		}
	}
	return out
}
