package defaults_test

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/defaults"
)

func TestWriteWorkspaceMkdirBlocked(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "boards"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := defaults.WriteWorkspace(root); err == nil {
		t.Fatal("expected mkdir fail")
	}
}
