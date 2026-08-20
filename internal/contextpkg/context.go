package contextpkg

import (
	"strings"

	"crm/internal/record"
)

type Bundle struct {
	Seed    *record.Record   `json:"seed"`
	Related []*record.Record `json:"related"`
}

// Assemble walks stable id relations from a seed record.
// Implements: SYS-REQ-260820-0TQX SW-REQ-260820-V48V
func Assemble(seed *record.Record, all []*record.Record) *Bundle {
	byID := map[string]*record.Record{}
	for _, rec := range all {
		byID[rec.ID] = rec
	}
	seen := map[string]bool{seed.ID: true}
	var related []*record.Record
	for _, id := range outgoing(seed) {
		if rec, ok := byID[id]; ok && !seen[id] {
			seen[id] = true
			related = append(related, rec)
		}
	}
	for _, rec := range all {
		if seen[rec.ID] {
			continue
		}
		for _, id := range outgoing(rec) {
			if id == seed.ID {
				seen[rec.ID] = true
				related = append(related, rec)
				break
			}
		}
	}
	return &Bundle{Seed: seed, Related: related}
}

// Implements: SYS-REQ-260820-0TQX
func outgoing(rec *record.Record) []string {
	var out []string
	for _, key := range []string{"company", "customer", "owner", "technical_owner"} {
		if s := rec.GetString(key); looksID(s) {
			out = append(out, s)
		}
	}
	for _, key := range []string{"people", "companies", "contacts", "related"} {
		for _, s := range record.StringSlice(rec.Get(key)) {
			if looksID(s) {
				out = append(out, s)
			}
		}
	}
	if rels, ok := rec.Get("relations").([]any); ok {
		for _, item := range rels {
			if m, ok := item.(map[string]any); ok {
				if t, ok := m["target"].(string); ok && looksID(t) {
					out = append(out, t)
				}
			}
		}
	}
	return out
}

// Implements: SYS-REQ-260820-0TQX
func looksID(s string) bool {
	s = strings.TrimSpace(s)
	return s != "" && (strings.Contains(s, "_") || strings.Contains(s, "-"))
}
