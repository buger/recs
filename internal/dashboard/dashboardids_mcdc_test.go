package dashboard

import (
	"os"
	"path/filepath"
	"testing"
)

// Verifies: SW-REQ-260821-E5V8 SW-REQ-260820-NA06
func TestDashboardIDsYmlAndSkip(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "dashboards")
	if err := os.MkdirAll(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skip.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "home.yml"), []byte("name: Home\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ids := dashboardIDs(root)
	if len(ids) != 1 || ids[0] != "home" {
		t.Fatalf("yml id: %v", ids)
	}
	if ids := dashboardIDs(t.TempDir()); len(ids) != 0 {
		t.Fatalf("missing dashboards dir: %v", ids)
	}
}
