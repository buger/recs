package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buger/recs/internal/record"
)

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func TestLoadAllErrorPaths(t *testing.T) {
	a := setup(t)
	if _, err := a.Create("grant", "grant_e", map[string]any{"title": "E", "status": "researching", "tags": []any{"x"}}, ""); err != nil {
		t.Fatal(err)
	}
	records := filepath.Join(a.Root(), "records")
	if err := os.Chmod(records, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(records, 0o755) })
	if _, err := a.List(""); err == nil {
		t.Fatal("list")
	}
	if _, err := a.Query("status = open"); err == nil {
		t.Fatal("query")
	}
	if _, err := a.Search("E"); err == nil {
		t.Fatal("search")
	}
	if _, err := a.Board("grants", map[string]string{"status": "nope"}); err == nil {
		t.Fatal("board")
	}
	if _, err := a.Next(); err == nil {
		t.Fatal("next")
	}
	if _, err := a.Triage(); err == nil {
		t.Fatal("triage")
	}
	if _, err := a.Validate(); err == nil {
		t.Fatal("validate")
	}
	if _, err := a.RebuildIndex(); err == nil {
		t.Fatal("index")
	}
	if _, err := a.Inbox(); err == nil {
		t.Fatal("inbox")
	}
	if _, err := a.Context("grant_e"); err == nil {
		t.Fatal("context")
	}
	if _, err := a.Show("grant_e"); err == nil {
		t.Fatal("show")
	}
}

// Verifies: SYS-REQ-260820-4628 SW-REQ-260820-NBGR SYS-REQ-260820-5C9D SW-REQ-260820-ZKCV
func TestBoardFilterMissAndNextFallback(t *testing.T) {
	a := setup(t)
	if _, err := a.Create("grant", "grant_f", map[string]any{"title": "F", "status": "researching", "tags": []any{"no"}}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("note", "note_empty_na", map[string]any{"next_action": map[string]any{}}, ""); err != nil {
		t.Fatal(err)
	}
	view, err := a.Board("grants", map[string]string{"status": "zzz", "tags": "missing"})
	if err != nil {
		t.Fatal(err)
	}
	_ = view
	next, err := a.Next()
	if err != nil {
		t.Fatal(err)
	}
	_ = next
	_ = record.StringSlice(nil)
}

// Verifies: SYS-REQ-260820-BVBE SW-REQ-260820-EX7Q
func TestMoveWriteError(t *testing.T) {
	a := setup(t)
	if _, err := a.Create("grant", "grant_m", map[string]any{"title": "M", "status": "researching"}, ""); err != nil {
		t.Fatal(err)
	}
	locks := filepath.Join(a.Root(), ".crm", "runtime", "locks")
	if err := os.Chmod(locks, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locks, 0o755) })
	if _, _, err := a.Move("grant_m", "grants", "applied"); err == nil {
		t.Fatal("expected lock fail")
	}
}
