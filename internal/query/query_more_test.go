package query

import (
	"testing"

	"crm/internal/record"
)

func TestParseAndMatchAllOps(t *testing.T) {
	if _, err := Parse(""); err == nil {
		t.Fatal("empty")
	}
	if _, err := Parse("????"); err == nil {
		t.Fatal("invalid")
	}
	rec := &record.Record{Fields: map[string]any{
		"status": "open", "n": 5, "tags": []any{"go", "crm"}, "title": "Hello",
	}}
	cases := []struct {
		expr string
		ok   bool
	}{
		{`status = open`, true},
		{`status != closed`, true},
		{`n < 10`, true},
		{`n > 1`, true},
		{`n <= 5`, true},
		{`n >= 5`, true},
		{`title contains ell`, true},
		{`tags in [go, rust]`, true},
		{`status = closed`, false},
		{`n < 1`, false},
		{`title contains zz`, false},
		{`tags in [rust]`, false},
	}
	for _, tc := range cases {
		got, err := Filter([]*record.Record{rec}, tc.expr)
		if err != nil {
			t.Fatalf("%s: %v", tc.expr, err)
		}
		if tc.ok && len(got) != 1 || !tc.ok && len(got) != 0 {
			t.Fatalf("%s => %d want ok=%v", tc.expr, len(got), tc.ok)
		}
	}
	if !Match(rec, nil) {
		t.Fatal("empty clauses")
	}
	if matchClause(rec, Clause{Op: "nope"}) {
		t.Fatal("unknown op")
	}
	if !equalish([]string{"Open"}, "open") || !equalish("open", "OPEN") {
		t.Fatal("equalish")
	}
	parts := splitClauses(`status = "open x" extra`)
	if len(parts) < 1 {
		t.Fatal(parts)
	}
	if _, err := parseClause("status = x"); err != nil {
		t.Fatal(err)
	}
	if _, err := parseClause(" = x"); err == nil {
		t.Fatal("missing field")
	}
	if _, err := parseClause("nolex"); err == nil {
		t.Fatal("invalid clause")
	}
	if len(splitList(`["a", "b",]`)) != 2 {
		t.Fatal(splitList(`["a", "b",]`))
	}
	if Search(nil, "   ") != nil {
		t.Fatal("empty search")
	}
	recs := []*record.Record{
		{ID: "a", Path: "a.md", Fields: map[string]any{"title": "Alpha"}, Body: "hello"},
		{ID: "b", Path: "b.md", Fields: nil, Body: "zzz"},
	}
	if len(Search(recs, "hello")) != 1 || len(Search(recs, "hello missing")) != 0 {
		t.Fatal("search")
	}
}

func TestParseQuotedAndNeq(t *testing.T) {
	cs, err := Parse(`title = "Hello World" status != closed`)
	if err != nil || len(cs) != 2 {
		t.Fatalf("%v %#v", err, cs)
	}
}
