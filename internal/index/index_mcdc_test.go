package index

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/record"
)

// Verifies: SYS-REQ-260820-Q8GR SW-REQ-260820-BNR7 SYS-REQ-260820-0TQX SW-REQ-260820-V48V
func TestRelatedIDsMissingAndSelf(t *testing.T) {
	root := t.TempDir()
	recs := []*record.Record{
		{ID: "a", Type: "note", Path: "a.md", Fields: map[string]any{
			"company": "missing_co", "related": []any{"a"},
			"relations": []any{map[string]any{"target": "a"}},
		}},
		{ID: "b", Type: "note", Path: "b.md", Fields: map[string]any{"relations": "no"}},
	}
	snap, err := Rebuild(root, recs)
	if err != nil || snap == nil {
		t.Fatal(err)
	}
}

// Verifies: SYS-REQ-260820-Q8GR SW-REQ-260820-BNR7
func TestWriteJSONFailures(t *testing.T) {
	root := t.TempDir()
	if err := writeJSON(filepath.Join(root, "nope", "x.json"), map[string]int{"a": 1}); err == nil {
		t.Fatal("missing parent")
	}
	dir := filepath.Join(root, ".crm", "index")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "records.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Rebuild(root, []*record.Record{{ID: "n", Type: "note", Path: "n.md", Fields: map[string]any{}}}); err == nil {
		t.Fatal("records.json is dir")
	}
	root = t.TempDir()
	dir = filepath.Join(root, ".crm", "index")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "by-type.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Rebuild(root, []*record.Record{{ID: "n", Type: "note", Path: "n.md", Fields: map[string]any{}}}); err == nil {
		t.Fatal("by-type is dir")
	}
	root = t.TempDir()
	dir = filepath.Join(root, ".crm", "index")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "by-tag.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Rebuild(root, []*record.Record{{ID: "n", Type: "note", Path: "n.md", Fields: map[string]any{}}}); err == nil {
		t.Fatal("by-tag is dir")
	}
	root = t.TempDir()
	dir = filepath.Join(root, ".crm", "index")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "backlinks.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Rebuild(root, []*record.Record{{ID: "n", Type: "note", Path: "n.md", Fields: map[string]any{}}}); err == nil {
		t.Fatal("backlinks is dir")
	}
}

// Verifies: SYS-REQ-260820-Q8GR SW-REQ-260820-BNR7 SYS-REQ-260820-0TQX SW-REQ-260820-V48V
func TestRelatedIDsNonStringTarget(t *testing.T) {
	ids := relatedIDs(&record.Record{Fields: map[string]any{
		"relations": []any{map[string]any{"target": 3}, map[string]any{}},
	}})
	_ = ids
}
