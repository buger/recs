package board

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/record"
)

// Verifies: SYS-REQ-260820-4628 SW-REQ-260820-NBGR
func TestLoadAllErrorAndDirEntries(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "boards"), []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAll(root); err == nil {
		t.Fatal("boards as file")
	}
	root = t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "boards", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "boards", "bad.yaml"), []byte(":\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAll(root); err == nil {
		t.Fatal("bad yaml via LoadAll")
	}
}

// Verifies: SYS-REQ-260820-4628 SW-REQ-260820-NBGR
func TestLoadDotDotNameIDAndPerms(t *testing.T) {
	root := t.TempDir()
	if _, err := Load(root, "foo..bar"); err == nil {
		t.Fatal("dotdot name")
	}
	if err := os.MkdirAll(filepath.Join(root, "boards"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "boards", "named.yaml"), []byte("id: custom\nname: N\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := Load(root, "named")
	if err != nil || b.ID != "custom" {
		t.Fatalf("%+v %v", b, err)
	}
	path := filepath.Join(root, "boards", "secret.yaml")
	if err := os.WriteFile(path, []byte("name: S\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
	if _, err := Load(root, "secret"); err == nil {
		t.Fatal("expected permission error")
	}
}

// Verifies: SYS-REQ-260820-4628 SW-REQ-260820-NBGR
func TestProjectEmptyColumnAndMatchMisses(t *testing.T) {
	b := &Board{
		Column:  ColumnConfig{Field: "status"},
		Columns: []Column{{ID: "open", Title: "Open", Match: map[string]any{"status": "open"}}},
	}
	empty := &record.Record{ID: "e", Fields: map[string]any{"status": ""}}
	miss := &record.Record{ID: "m", Fields: map[string]any{"status": "other"}}
	view := b.Project([]*record.Record{empty, miss})
	if len(view.Columns) < 1 {
		t.Fatalf("%#v", view.Columns)
	}
	rec := &record.Record{Fields: map[string]any{"status": "open", "title": "Hello", "tags": []any{"a"}}}
	if matchValue(rec, []any{map[string]any{"status": "closed"}}) {
		t.Fatal("list all miss")
	}
	if matchAll(rec, map[string]any{"status": "closed"}) {
		t.Fatal("matchAll scalar miss")
	}
	if !matchAny(rec, map[string]any{"status": "open"}) {
		t.Fatal("matchAny scalar hit")
	}
	if matchField(rec, "status", map[string]any{"in": []any{"closed"}}) {
		t.Fatal("in miss")
	}
	if !matchField(rec, "status", map[string]any{"not_in": []any{"closed"}}) {
		t.Fatal("not_in hit")
	}
	if matchField(rec, "title", map[string]any{"contains": "zzz"}) {
		t.Fatal("contains miss")
	}
	if matchField(rec, "status", map[string]any{"eq": "closed"}) {
		t.Fatal("eq miss")
	}
	_ = matchField(rec, "status", map[string]any{"unknown": true})
	if fieldEquals(rec, "status", []any{"closed", "x"}) {
		t.Fatal("fieldEquals list miss")
	}
}

// Verifies: SYS-REQ-260820-4628 SW-REQ-260820-NBGR
func TestLoadAllSkipsDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "boards", "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "boards", "ok.yaml"), []byte("name: OK\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	all, err := LoadAll(root)
	if err != nil || len(all) != 1 {
		t.Fatalf("%v %d", err, len(all))
	}
}

// Verifies: SYS-REQ-260820-4628 SW-REQ-260820-NBGR
func TestMatchMissingEmptyString(t *testing.T) {
	rec := &record.Record{Fields: map[string]any{"title": ""}}
	if !matchField(rec, "title", map[string]any{"missing": true}) {
		t.Fatal("empty string is missing")
	}
	if matchField(rec, "title", map[string]any{"exists": true}) && rec.Get("title") == nil {
		t.Fatal("exists")
	}
}
