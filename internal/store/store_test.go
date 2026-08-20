package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/app"
	"crm/internal/record"
)

// Verifies: SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2
func TestInitCreatesLayout(t *testing.T) {
	root := t.TempDir()
	a := app.OpenOrCWD(root)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"crm.yaml", "records", "boards", "inbox", "attachments", "templates", ".crm/index", ".crm/cache", ".crm/runtime"} {
		if _, err := os.Stat(filepath.Join(a.Root(), p)); err != nil {
			t.Fatalf("missing %s: %v", p, err)
		}
	}
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4
func TestCreateShowPatch(t *testing.T) {
	a := initApp(t)
	rec, err := a.Create("grant", "", map[string]any{"title": "Solana MC/DC", "status": "researching"}, "notes")
	if err != nil {
		t.Fatal(err)
	}
	got, err := a.Show(rec.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != "grant" || got.GetString("status") != "researching" {
		t.Fatalf("unexpected %#v", got.Fields)
	}
	res, err := a.Set(rec.ID, "status", "preparing")
	if err != nil {
		t.Fatal(err)
	}
	if res.Record.GetString("status") != "preparing" {
		t.Fatal(res.Record.GetString("status"))
	}
	_, err = a.Patch(rec.ID, map[string]any{"status": "applied"}, nil, "sha256:dead")
	if err == nil {
		t.Fatal("expected conflict")
	}
}

// Verifies: SYS-REQ-260820-7WT4 SW-REQ-260820-9C5Z SYS-REQ-260820-BVBE SW-REQ-260820-EX7Q INT-REQ-260820-JRWN SYS-REQ-260820-4628 SW-REQ-260820-NBGR
func TestMoveKeepsPath(t *testing.T) {
	a := initApp(t)
	rec, err := a.Create("grant", "grant_demo", map[string]any{"title": "Demo", "status": "researching"}, "")
	if err != nil {
		t.Fatal(err)
	}
	old := rec.Path
	moved, path, err := a.Move(rec.ID, "grants", "applied")
	if err != nil {
		t.Fatal(err)
	}
	if path != old || moved.Path != old {
		t.Fatalf("relocated %s -> %s", old, moved.Path)
	}
	if moved.GetString("status") != "applied" {
		t.Fatalf("status %s", moved.GetString("status"))
	}
	if _, err := os.Stat(old); err != nil {
		t.Fatal(err)
	}
}

func initApp(t *testing.T) *app.App {
	t.Helper()
	root := t.TempDir()
	a := app.OpenOrCWD(root)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	return a
}

var _ = record.Record{}
