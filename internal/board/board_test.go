package board

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/record"
)

func writeBoard(t *testing.T, root, name, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "boards"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "boards", name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Verifies: SYS-REQ-260820-4628 SW-REQ-260820-NBGR
func TestLoadAndProject(t *testing.T) {
	root := t.TempDir()
	if got, err := LoadAll(root); err != nil || got != nil {
		t.Fatalf("%v %#v", err, got)
	}
	writeBoard(t, root, "notes.txt", "nope")
	writeBoard(t, root, "subdir", "")
	os.Mkdir(filepath.Join(root, "boards", "subdir"), 0o755)
	writeBoard(t, root, "grants.yml", "name: Grants\ncolumns:\n  - id: open\n    title: Open\n    match:\n      status: open\n  - id: done\n    title: Done\n    on_drop:\n      set:\n        status: done\n        closed_at: $now\n        extra: 1\n      remove: [owner]\n")
	if _, err := Load(root, ""); err == nil {
		t.Fatal("empty name")
	}
	if _, err := Load(root, "../x"); err == nil {
		t.Fatal("slash")
	}
	if _, err := Load(root, "missing"); err == nil {
		t.Fatal("missing")
	}
	all, err := LoadAll(root)
	if err != nil || len(all) != 1 {
		t.Fatalf("%v %d", err, len(all))
	}
	b, err := Load(root, "grants")
	if err != nil || b.Column.Field != "status" || b.ID != "grants" {
		t.Fatalf("%+v %v", b, err)
	}
	open := &record.Record{ID: "g1", Fields: map[string]any{"status": "open", "type": "grant"}}
	done := &record.Record{ID: "g2", Fields: map[string]any{"status": "done"}}
	other := &record.Record{ID: "x", Fields: map[string]any{"status": "other"}}
	if !b.Matches(open) {
		t.Fatal("match nil")
	}
	b.Match = map[string]any{"type": "grant"}
	if !b.Matches(open) || b.Matches(done) {
		t.Fatal("match type")
	}
	b.Filters = map[string]any{"status": "open"}
	if !b.Matches(open) || b.Matches(&record.Record{Fields: map[string]any{"type": "grant", "status": "x"}}) {
		t.Fatal("filters")
	}
	b.Filters = nil
	if b.ColumnOf(open) != "open" || b.ColumnOf(other) != "other" {
		t.Fatal(b.ColumnOf(open), b.ColumnOf(other))
	}
	if _, ok := b.ColumnByID("nope"); ok {
		t.Fatal("unknown col")
	}
	view := b.Project([]*record.Record{open, done, other})
	if len(view.Columns) < 2 {
		t.Fatalf("%#v", view.Columns)
	}
	if err := b.ApplyMove(open, "nope"); err == nil {
		t.Fatal("unknown move")
	}
	if err := b.ApplyMove(open, "done"); err != nil {
		t.Fatal(err)
	}
	if open.GetString("status") != "done" || open.Get("closed_at") == nil {
		t.Fatal(open.Fields)
	}
	if err := b.ApplyMove(done, "open"); err != nil {
		t.Fatal(err)
	}
}

// Verifies: SYS-REQ-260820-4628 SW-REQ-260820-NBGR
func TestMatchHelpers(t *testing.T) {
	rec := &record.Record{Fields: map[string]any{"status": "open", "tags": []any{"a", "b"}, "title": "Hello"}}
	if !matchValue(rec, map[string]any{"all": []any{map[string]any{"status": "open"}}}) {
		t.Fatal("all")
	}
	if matchValue(rec, map[string]any{"all": []any{map[string]any{"status": "x"}}}) {
		t.Fatal("all fail")
	}
	if !matchValue(rec, map[string]any{"any": []any{map[string]any{"status": "x"}, map[string]any{"status": "open"}}}) {
		t.Fatal("any")
	}
	if matchValue(rec, map[string]any{"any": []any{map[string]any{"status": "x"}}}) {
		t.Fatal("any fail")
	}
	if !matchValue(rec, []any{map[string]any{"status": "open"}}) {
		t.Fatal("list all")
	}
	if matchValue(rec, "nope") {
		t.Fatal("default")
	}
	if !matchAll(rec, map[string]any{"status": "open"}) || matchAll(rec, []any{map[string]any{"status": "x"}}) {
		t.Fatal("matchAll")
	}
	if matchAny(rec, map[string]any{"status": "x"}) || matchAny(rec, []any{}) {
		t.Fatal("matchAny")
	}
	if !matchField(rec, "status", map[string]any{"in": []any{"open", "x"}}) {
		t.Fatal("in")
	}
	if matchField(rec, "status", map[string]any{"not_in": []any{"open"}}) {
		t.Fatal("not_in")
	}
	if !matchField(rec, "title", map[string]any{"contains": "ell"}) {
		t.Fatal("contains")
	}
	if !matchField(rec, "status", map[string]any{"eq": "open"}) {
		t.Fatal("eq")
	}
	if !matchField(rec, "status", map[string]any{"exists": true}) || matchField(rec, "nope", map[string]any{"exists": true}) {
		t.Fatal("exists")
	}
	if !matchField(rec, "nope", map[string]any{"missing": true}) || matchField(rec, "status", map[string]any{"missing": true}) {
		t.Fatal("missing")
	}
	if !matchField(rec, "status", []any{"open"}) || !matchField(rec, "status", "open") {
		t.Fatal("eq list")
	}
	if !fieldEquals(rec, "tags", []any{"a", "b"}) || !inList(rec.Get("tags"), []string{"A"}) {
		t.Fatal("inList")
	}
	if inList(nil, "z") || !inList("z", "z") {
		t.Fatal("inList scalar")
	}
	if expandValue(1) != 1 || expandValue("hi") != "hi" {
		t.Fatal("expand")
	}
	if expandValue("$now") == "$now" {
		t.Fatal("now")
	}
	writeBoard(t, t.TempDir(), "bad.yaml", ":\n")
}
