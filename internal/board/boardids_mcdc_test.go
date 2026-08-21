package board

import (
	"os"
	"path/filepath"
	"testing"
)

// Verifies: SW-REQ-260821-E5V8
func TestBoardIDsErrorAndDir(t *testing.T) {
	missing := t.TempDir()
	if ids := boardIDs(missing); len(ids) != 0 {
		t.Fatalf("missing boards dir: %v", ids)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "boards", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "boards", "ok.yaml"), []byte("name: OK\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "boards", "also.yml"), []byte("name: Y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ids := boardIDs(root)
	if len(ids) < 2 {
		t.Fatalf("expected yaml and yml ids, got %v", ids)
	}
}

// Verifies: SW-REQ-260821-E5V8
func TestBoardIDsBoardsIsFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "boards"), []byte("not-a-dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	if ids := boardIDs(root); len(ids) != 0 {
		t.Fatalf("boards as file: %v", ids)
	}
}
