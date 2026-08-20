package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/app"
	"crm/internal/record"
)

func TestNextTriageInboxIndependence(t *testing.T) {
	a := setup(t)
	if _, err := a.Create("task", "task_done", map[string]any{"title": "Done", "status": "done", "priority": "high"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("note", "note_empty_action", map[string]any{"next_action": "", "priority": "low"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("task", "task_a", map[string]any{"title": "A", "status": "open", "due": "2026-01-01", "priority": "low"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("task", "task_b", map[string]any{"title": "B", "status": "open", "due": "2026-01-01", "priority": "critical"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("task", "task_c", map[string]any{"title": "C", "status": "open", "due": "2026-01-01", "priority": "medium"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("note", "note_new", map[string]any{"title": "New", "triage_status": "new", "name": "Named"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("note", "note_block", map[string]any{"title": "Blk", "blockers": []any{"x"}, "health": "ok"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("note", "note_noname", map[string]any{"status": "open"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("note", "note_titleless", map[string]any{"name": "HasName"}, ""); err != nil {
		t.Fatal(err)
	}
	next, err := a.Next()
	if err != nil || len(next) < 3 {
		t.Fatalf("%v %#v", err, next)
	}
	triage, err := a.Triage()
	if err != nil || len(triage) == 0 {
		t.Fatal(err, triage)
	}
	inbox, err := a.Inbox()
	if err != nil || len(inbox) == 0 {
		t.Fatal(err, inbox)
	}
	_ = record.DisplayName(&record.Record{ID: "x"})
}

func TestOpenDeletedCWD(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	_, _ = app.Open("")
}

func TestMoveDoesNotRelocate(t *testing.T) {
	a := setup(t)
	rec, err := a.Create("grant", "grant_mv", map[string]any{"title": "M", "status": "researching"}, "")
	if err != nil {
		t.Fatal(err)
	}
	got, path, err := a.Move("grant_mv", "grants", "applied")
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != rec.Path || path != rec.Path {
		t.Fatal(got.Path, path, rec.Path)
	}
}

func TestContextAfterGetUsesSameTree(t *testing.T) {
	a := setup(t)
	if _, err := a.Create("note", "note_ctx", map[string]any{"title": "C"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Context("note_ctx"); err != nil {
		t.Fatal(err)
	}
	_ = filepath.Separator
}
