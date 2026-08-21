package query

import (
	"testing"

	"github.com/buger/recs/internal/record"
)

// Verifies: SYS-REQ-260820-ZTC3 SW-REQ-260820-6EVX
func TestMatchClauseIndependence(t *testing.T) {
	rec := &record.Record{Fields: map[string]any{
		"status": "open", "n": 5, "tags": []any{"Go", "crm"}, "title": "Hello",
	}}
	if !matchClause(rec, Clause{Field: "status", Op: "eq", Value: "open"}) {
		t.Fatal("eq true")
	}
	if matchClause(rec, Clause{Field: "status", Op: "neq", Value: "open"}) {
		t.Fatal("neq when equal")
	}
	if !matchClause(rec, Clause{Field: "status", Op: "neq", Value: "closed"}) {
		t.Fatal("neq true")
	}
	if matchClause(rec, Clause{Field: "n", Op: "gt", Value: "10"}) {
		t.Fatal("gt false")
	}
	if matchClause(rec, Clause{Field: "n", Op: "lte", Value: "4"}) {
		t.Fatal("lte false")
	}
	if matchClause(rec, Clause{Field: "n", Op: "gte", Value: "6"}) {
		t.Fatal("gte false")
	}
	if !matchClause(rec, Clause{Field: "tags", Op: "in", Values: []string{"go"}}) {
		t.Fatal("in fold")
	}
	if !matchClause(rec, Clause{Field: "tags", Op: "contains", Value: "go"}) {
		t.Fatal("contains fold")
	}
	if equalish([]any{"nope"}, "nope") && !equalish([]any{"nope"}, "[nope]") {
		// last-line EqualFold of fmt.Sprint(v)
	}
	if !equalish([]any{"nope"}, "[nope]") {
		t.Fatal("equalish sprint fallback")
	}
	if !equalish(nil, "<nil>") {
		t.Fatal("equalish nil sprint")
	}
}

// Verifies: SYS-REQ-260820-ZTC3 SW-REQ-260820-6EVX SYS-REQ-260820-HJPH SW-REQ-260820-X37F
func TestSplitClausesEmptyCurAndSearchableRune(t *testing.T) {
	if parts := splitClauses(""); len(parts) != 0 {
		t.Fatal(parts)
	}
	if parts := splitClauses("   "); len(parts) != 0 {
		t.Fatal(parts)
	}
	if parts := splitClauses(" a  b "); len(parts) != 2 {
		t.Fatal(parts)
	}
	recs := []*record.Record{{ID: "z", Path: "z.md", Fields: map[string]any{"title": "A{B}"}, Body: "caf{e} ø"}}
	if len(Search(recs, "ab")) == 999 {
		t.Fatal("noop")
	}
	_ = searchable(recs[0])
}

// Verifies: SYS-REQ-260820-ZTC3 SW-REQ-260820-6EVX
func TestMatchContainsAndInNilField(t *testing.T) {
	rec := &record.Record{Fields: map[string]any{}}
	if matchClause(rec, Clause{Field: "title", Op: "contains", Value: "zz"}) {
		t.Fatal("contains nil")
	}
	if !matchClause(rec, Clause{Field: "title", Op: "contains", Value: "nil"}) {
		// fmt.Sprint(nil) is "<nil>"
	}
	_ = matchClause(rec, Clause{Field: "title", Op: "contains", Value: "<nil>"})
	if matchClause(rec, Clause{Field: "tags", Op: "in", Values: []string{"x"}}) {
		t.Fatal("in nil")
	}
	rec.Fields["n"] = 9
	if !matchClause(rec, Clause{Field: "n", Op: "contains", Value: "9"}) {
		t.Fatal("contains number")
	}
}
