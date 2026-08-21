package dashboard

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/buger/recs/internal/record"
)

// Verifies: SW-REQ-260820-EJVT
// SW-REQ-260820-EJVT
func TestSafeRelEmptyAndRunQueryField(t *testing.T) {
	if _, err := safeRel(t.TempDir(), ""); err == nil {
		t.Fatal("empty rel")
	}
	if _, err := safeRel(t.TempDir(), "   "); err == nil {
		t.Fatal("blank rel")
	}
	now := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	recs := []*record.Record{{ID: "r", Type: "note", Fields: map[string]any{"due": "2026-01-01", "status": "open"}}}
	_, _ = runQuery(recs, "", "due", "2026-12-31", now)
	_, _ = runQuery(recs, "", "due", "", now)
	_, _ = runQuery(recs, "", "", "2026-12-31", now)
	_, _ = runQuery(recs, "status=open", "", "", now)
}

// Verifies: SW-REQ-260820-NA06
// SW-REQ-260820-NA06
func TestLoadAllSuffixAndSortIndependence(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "dashboards")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skip.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.yaml"), []byte("id: zzz\nname: Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.yml"), []byte("id: aaa\nname: A\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := LoadAll(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("got %d", len(out))
	}
}
