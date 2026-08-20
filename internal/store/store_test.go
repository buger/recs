package store_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crm/internal/app"
	"crm/internal/record"
	"crm/internal/store"
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

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y SYS-REQ-260820-7WT4
func TestInboxAndTemplate(t *testing.T) {
	a := initApp(t)
	rec, err := a.Create("grant", "grant_tpl2", map[string]any{"title": "T"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rec.Body, "Opportunity") {
		t.Fatalf("template not applied: %q", rec.Body)
	}
	inbox := filepath.Join(a.Root(), "inbox", "loose.md")
	if err := os.WriteFile(inbox, []byte("---\nid: loose_1\ntype: note\n---\n\n# Loose\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := a.Show("loose_1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "loose_1" {
		t.Fatalf("%#v", got)
	}
}

func TestConflictAndNow(t *testing.T) {
	a := initApp(t)
	rec, err := a.Create("grant", "grant_now", map[string]any{"title": "Now", "status": "researching"}, "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.Patch(rec.ID, map[string]any{"status": "applied"}, nil, "sha256:dead")
	var conf *store.ConflictError
	if !errors.As(err, &conf) || conf.Expected == "" || conf.Current == "" {
		t.Fatalf("conflict shape: %v", err)
	}
	res, err := a.Patch(rec.ID, map[string]any{"applied_at": "now"}, nil, rec.Version())
	if err != nil {
		t.Fatal(err)
	}
	got := res.Record.GetString("applied_at")
	if got == "" || got == "now" {
		t.Fatalf("now not expanded: %q", got)
	}
}
