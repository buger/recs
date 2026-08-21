package app

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/record"
)

// Verifies: SW-REQ-260821-8C2C
// SW-REQ-260821-8C2C
func TestRelationsOfIndependence(t *testing.T) {
	if got := relationsOf(nil); len(got) != 0 {
		t.Fatalf("nil rec: %#v", got)
	}
	rec := &record.Record{ID: "n", Type: "note", Fields: map[string]any{
		"relations": []any{"not-a-map", map[string]any{"type": "a", "target": "b"}},
	}}
	got := relationsOf(rec)
	if len(got) != 1 || got[0].Type != "a" {
		t.Fatalf("%#v", got)
	}
}

// Verifies: SW-REQ-260821-8C2C
// SW-REQ-260821-8C2C
func TestStoreConfinedEscape(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Dir(root)
	if err := storeConfined(root, parent); err == nil {
		t.Fatal("parent should escape")
	}
	outside := filepath.Join(parent, "outside-sibling")
	if err := storeConfined(root, outside); err == nil {
		t.Fatal("sibling should escape")
	}
	if err := storeConfined(root, filepath.Join(root, "ok")); err != nil {
		t.Fatal(err)
	}
}

// Verifies: SW-REQ-260821-8C2C
// SW-REQ-260821-8C2C
func TestRunGitEmptyStderr(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "git")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	if _, err := runGit(dir, "status"); err == nil {
		t.Fatal("expected git fail")
	}
}
