package store

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/record"
)

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2
func TestCreateBranchesAndLoadAll(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(&record.Record{Fields: map[string]any{}}); err == nil {
		t.Fatal("empty type")
	}
	rec := &record.Record{Type: "note", Fields: map[string]any{"title": "T", "created_at": "2020-01-01T00:00:00Z"}, Path: filepath.Join(s.Root, "records", "notes", "note_preset.md")}
	if err := s.Create(rec); err != nil {
		t.Fatal(err)
	}
	rec2 := &record.Record{Type: "note", ID: "", Fields: map[string]any{"name": "Named"}}
	if err := s.Create(rec2); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.Root, "records", "notes", "skip.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.Root, "records", "notes", "noid.md"), []byte("---\ntype: note\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.Root, "inbox", "dup.md"), []byte("---\nid: note_preset\ntype: note\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	recs, err := s.LoadAll()
	if err != nil || len(recs) < 2 {
		t.Fatal(err, len(recs))
	}
	empty := &Store{Root: t.TempDir()}
	got, err := empty.LoadAll()
	if err != nil || got != nil {
		t.Fatal(err, got)
	}
	s.ApplyTemplate(nil)
	s.ApplyTemplate(&record.Record{})
	s.ApplyTemplate(&record.Record{Type: "grant", Fields: map[string]any{"status": ""}})
	if err := confinedToRoot(s.Root, filepath.Join(s.Root, "..", "escape.md")); err == nil {
		t.Fatal("escape")
	}
	if err := confinedToRoot(s.Root, filepath.Join(s.Root, "records", "notes", "ok.md")); err != nil {
		t.Fatal(err)
	}
	var nilLock *fileLock
	nilLock.Close()
	(&fileLock{}).Close()
	if expandNow(1) != 1 || expandNow("$now") == "$now" || expandNow("now") == "now" {
		t.Fatal("expandNow")
	}
	if err := s.Create(&record.Record{Type: "note", ID: rec.ID, Fields: map[string]any{"type": "note"}}); err == nil {
		t.Fatal("exists")
	}
	if err := s.Write(&record.Record{ID: "x"}); err == nil {
		t.Fatal("empty path")
	}
	res, err := s.Patch(rec.ID, map[string]any{"extra": 1}, []string{"name"}, "")
	if err != nil || res == nil {
		t.Fatal(err)
	}
}

// Verifies: SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2 SYS-REQ-260820-9J7C
func TestFindRootAndBadFiles(t *testing.T) {
	if _, err := FindRoot(t.TempDir()); err == nil {
		t.Fatal("missing root")
	}
	s := &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(s.Root, "records", "notes"), 0755); err != nil { t.Fatal(err) }
	if err := os.WriteFile(filepath.Join(s.Root, "records", "notes", "bad.md"), []byte("---\n:\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.LoadAll(); err == nil {
		t.Fatal("bad md")
	}
	if _, err := s.Get("nope"); err == nil {
		t.Fatal("missing get")
	}
	if err := os.WriteFile(filepath.Join(s.Root, "templates", "note.md"), []byte("---\n:\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s.ApplyTemplate(&record.Record{Type: "note", Fields: map[string]any{}})
	if err := os.WriteFile(filepath.Join(s.Root, "crm.yaml"), []byte(":\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec := &record.Record{Type: "grant", ID: "grant_z", Fields: map[string]any{"status": "open"}}
	_ = s.checkEnum(rec, map[string]any{"status": "open"})
}
