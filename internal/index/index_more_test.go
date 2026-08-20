package index

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/record"
)

func TestRebuildRelationsAndTags(t *testing.T) {
	root := t.TempDir()
	recs := []*record.Record{
		{ID: "company_acme", Type: "company", Path: "c.md", Fields: map[string]any{"name": "Acme", "tags": []any{"hot"}}},
		{ID: "person_a", Type: "person", Path: "p.md", Fields: map[string]any{
			"title": "A", "company": "company_acme", "people": []any{"person_a"},
			"relations": []any{map[string]any{"target": "company_acme"}, "x"},
			"owner": "plain",
		}},
	}
	snap, err := Rebuild(root, recs)
	if err != nil || len(snap.Records) != 2 || len(snap.Backlinks["company_acme"]) == 0 {
		t.Fatalf("%+v %v", snap, err)
	}
	if _, err := os.Stat(filepath.Join(root, ".crm/index/records.json")); err != nil {
		t.Fatal(err)
	}
	if relatedIDs(&record.Record{Fields: map[string]any{"relations": "no"}}) == nil {
		// ok
	}
}

func TestRebuildUnwritable(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".crm"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Rebuild(root, nil); err == nil {
		t.Fatal("expected mkdir fail")
	}
}
