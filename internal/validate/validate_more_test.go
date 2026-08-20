package validate_test

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/record"
	"crm/internal/validate"
)

func TestCheckSchemaAbsentAndRequired(t *testing.T) {
	root := t.TempDir()
	res, err := validate.Check(root, nil)
	if err != nil || res.SchemaPresent || !res.OK {
		t.Fatalf("%+v %v", res, err)
	}
	if err := os.WriteFile(filepath.Join(root, "crm.yaml"), []byte("name: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err = validate.Check(root, nil)
	if err != nil || res.SchemaPresent {
		t.Fatalf("%+v %v", res, err)
	}
	if err := os.WriteFile(filepath.Join(root, "crm.yaml"), []byte(":\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := validate.Check(root, nil); err == nil {
		t.Fatal("bad yaml")
	}
	if err := os.WriteFile(filepath.Join(root, "crm.yaml"), []byte("types:\n  grant:\n    required: [title]\n    fields:\n      status:\n        enum: [open, closed]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	okRec := &record.Record{ID: "g1", Type: "grant", Fields: map[string]any{"title": "T", "status": "open"}}
	miss := &record.Record{ID: "g2", Type: "grant", Fields: map[string]any{"status": ""}}
	other := &record.Record{ID: "n1", Type: "note", Fields: map[string]any{}}
	res, err = validate.Check(root, []*record.Record{okRec, miss, other})
	if err != nil || res.OK || len(res.Violations) == 0 {
		t.Fatalf("%+v %v", res, err)
	}
	v := res.Violations[0]
	if v.String() == "" {
		t.Fatal("string")
	}
	plain := validate.Violation{Record: "r", Error: "e"}
	if plain.String() != "r: e" {
		t.Fatal(plain.String())
	}
}

func TestLoadConfigMissing(t *testing.T) {
	if _, err := validate.LoadConfig(t.TempDir()); err == nil {
		t.Fatal("expected missing")
	}
}
