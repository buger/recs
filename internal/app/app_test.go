package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buger/recs/internal/app"
	"github.com/buger/recs/internal/record"
)

func setup(t *testing.T) *app.App {
	t.Helper()
	a := app.OpenOrCWD(t.TempDir())
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	return a
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2
func TestOpenAndCreateVariants(t *testing.T) {
	if _, err := app.Open(t.TempDir()); err == nil {
		t.Fatal("no workspace")
	}
	a := setup(t)
	if _, err := app.Open(a.Root()); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Open(""); err == nil {
		// cwd may or may not be a workspace
	}
	_ = app.OpenOrCWD("")
	if a.Root() == "" {
		t.Fatal("root")
	}
	rec, err := a.Create("note", "", map[string]any{"name": "N"}, "")
	if err != nil || rec.Body == "" {
		t.Fatal(err)
	}
	rec2, err := a.Create("note", "note_x", nil, "body")
	if err != nil || rec2.ID != "note_x" {
		t.Fatal(err, rec2)
	}
	if _, err := a.Create("grant", "bad/id", nil, ""); err == nil {
		t.Fatal("bad id")
	}
}

// Verifies: SYS-REQ-260820-ZTC3 SW-REQ-260820-6EVX SYS-REQ-260820-HJPH SW-REQ-260820-X37F SYS-REQ-260820-4628 SW-REQ-260820-NBGR SYS-REQ-260820-5C9D SW-REQ-260820-ZKCV SYS-REQ-260820-DCG4 SW-REQ-260820-D5WE
func TestListQuerySearchBoardNextTriage(t *testing.T) {
	a := setup(t)
	if _, err := a.Create("grant", "grant_a", map[string]any{"title": "A", "status": "inbox", "priority": "high", "due": "2000-01-01", "health": "at_risk", "blockers": []any{"x"}, "tags": []any{"go"}}, "alpha"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("task", "task_open", map[string]any{"title": "Do", "status": "open", "priority": "critical", "due": "2099-01-01"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("note", "note_next", map[string]any{"next_action": map[string]any{"action": "Call", "date": "2001-01-01"}, "priority": "low"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("note", "note_na", map[string]any{"next_action": "Ping", "priority": "medium"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("note", "note_blank", map[string]any{"next_action": map[string]any{"date": "2002-01-01"}, "priority": "med"}, ""); err != nil {
		t.Fatal(err)
	}
	all, err := a.List("")
	if err != nil || len(all) < 4 {
		t.Fatal(err, len(all))
	}
	grants, err := a.List("grant")
	if err != nil || len(grants) != 1 {
		t.Fatal(err, len(grants))
	}
	got, err := a.Query(`status = inbox`)
	if err != nil || len(got) != 1 {
		t.Fatal(err, len(got))
	}
	if _, err := a.Query("???"); err == nil {
		t.Fatal("bad query")
	}
	if sr, err := a.Search("alpha"); err != nil || len(sr) != 1 {
		t.Fatal("search")
	}
	boards, err := a.ListBoards()
	if err != nil || len(boards) == 0 {
		t.Fatal(err, boards)
	}
	view, err := a.Board("grants", map[string]string{"status": "inbox"})
	if err != nil {
		t.Fatal(err)
	}
	_ = view
	if _, err := a.Board("missing", nil); err == nil {
		t.Fatal("missing board")
	}
	view, err = a.Board("grants", map[string]string{"tags": "go"})
	if err != nil {
		t.Fatal(err)
	}
	_ = view
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
	if err := os.WriteFile(filepath.Join(a.Root(), "inbox", "loose.md"), []byte("---\nid: loose\ntype: note\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	inbox, err = a.Inbox()
	if err != nil {
		t.Fatal(err)
	}
	foundLoose := false
	for _, rec := range inbox {
		if rec.ID == "loose" {
			foundLoose = true
		}
	}
	if !foundLoose {
		t.Fatal("inbox path")
	}
	if _, err := a.Show("grant_a"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Set("grant_a", "status", "researching"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Patch("grant_a", map[string]any{"status": "applied"}, nil, ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := a.Move("grant_a", "grants", "researching"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := a.Move("missing", "grants", "open"); err == nil {
		t.Fatal("move missing rec")
	}
	if _, _, err := a.Move("grant_a", "missing", "open"); err == nil {
		t.Fatal("move missing board")
	}
	if _, err := a.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.RebuildIndex(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Context("grant_a"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Context("nope"); err == nil {
		t.Fatal("missing context")
	}
	_ = record.DisplayName(&record.Record{ID: "x"})
}

