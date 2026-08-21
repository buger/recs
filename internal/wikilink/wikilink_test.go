package wikilink_test

import (
	"strings"
	"testing"

	"crm/internal/record"
	"crm/internal/wikilink"
)

// Verifies: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func TestMatchAndResolve(t *testing.T) {
	alice := &record.Record{ID: "person_alice", Type: "person", Fields: map[string]any{"name": "Alice Smith"}}
	acme := &record.Record{ID: "company_acme", Type: "company", Fields: map[string]any{"name": "Acme"}}
	recs := []*record.Record{alice, acme, nil}
	if wikilink.Match("", recs) != nil || wikilink.Match("nobody", recs) != nil {
		t.Fatal("empty/missing")
	}
	if got := wikilink.Match("Alice Smith", recs); got == nil || got.ID != "person_alice" {
		t.Fatal("name")
	}
	if got := wikilink.Match("person_alice", recs); got == nil {
		t.Fatal("id")
	}
	if got := wikilink.Match("alice", recs); got == nil {
		t.Fatal("fuzzy")
	}
	html := wikilink.Resolve("Talked with [[Alice Smith]] about [[Missing]] and [[Acme|The Co]].", recs)
	if !strings.Contains(html, `#/r/person_alice`) || !strings.Contains(html, "wikilink missing") || !strings.Contains(html, "The Co") {
		t.Fatal(html)
	}
	if len(wikilink.FindAll("[[A]] [[B|C]]")) != 2 {
		t.Fatal("findall")
	}
}
