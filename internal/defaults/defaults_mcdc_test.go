package defaults_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buger/recs/internal/defaults"
)

// Verifies: SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2
func TestWriteWorkspaceMkdirBlocked(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "boards"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := defaults.WriteWorkspace(root); err == nil {
		t.Fatal("expected mkdir fail")
	}
}
