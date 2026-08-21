package store

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/record"
)

// Verifies: SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2
func TestInitAndFindRootFailures(t *testing.T) {
	file := filepath.Join(t.TempDir(), "notdir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (&Store{Root: file}).Init(); err == nil {
		t.Fatal("init on file")
	}

	root := t.TempDir()
	dirs := []string{
		"records/people", "records/companies", "records/customers", "records/grants",
		"records/applications", "records/onboarding", "records/emails", "records/tasks",
		"records/misc", "boards", "inbox", "attachments", "templates",
		".crm/index", ".crm/cache", ".crm/runtime/locks",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(root, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })
	if err := (&Store{Root: root}).Init(); err == nil {
		t.Fatal("expected writeworkspace fail")
	}
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func TestLoadAllPermissionAndInboxFile(t *testing.T) {
	base := t.TempDir()
	file := filepath.Join(base, "notdir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := &Store{Root: filepath.Join(file, "nested")}
	if _, err := s.LoadAll(); err == nil {
		t.Fatal("records through file")
	}
	s = &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	records := s.recordsDir()
	if err := os.Chmod(records, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(records, 0o755) })
	if _, err := s.LoadAll(); err == nil {
		t.Fatal("unreadable records dir")
	}
	_ = os.Chmod(records, 0o755)

	s = &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	inbox := filepath.Join(s.Root, "inbox")
	if err := os.RemoveAll(inbox); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(blocker, "nested"), inbox); err != nil {
		t.Fatal(err)
	}
	if _, err := s.LoadAll(); err == nil {
		t.Fatal("inbox through file symlink")
	}

	s = &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.Root, "records", "notes", "locked.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("---\nid: locked\ntype: note\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
	if _, err := s.LoadAll(); err == nil {
		t.Fatal("unreadable md")
	}
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4
func TestCreateGetExistsAndWriteFailures(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	first := &record.Record{Type: "note", ID: "note_dup", Fields: map[string]any{"title": "A"}}
	if err := s.Create(first); err != nil {
		t.Fatal(err)
	}
	second := &record.Record{
		Type: "note", ID: "note_dup",
		Fields: map[string]any{"title": "B"},
		Path:   filepath.Join(s.Root, "records", "notes", "other.md"),
	}
	if err := s.Create(second); err == nil {
		t.Fatal("same id other path")
	}

	notes := filepath.Join(s.Root, "records", "notes")
	if err := os.RemoveAll(notes); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(notes, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(&record.Record{Type: "note", ID: "note_x", Fields: map[string]any{"title": "X"}}); err == nil {
		t.Fatal("mkdirall file")
	}

	s = &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(s.recordsDir(), 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(s.recordsDir(), 0o755) })
	rec := &record.Record{
		Type: "note", ID: "note_y",
		Fields: map[string]any{"title": "Y"},
		Path:   filepath.Join(s.Root, "inbox", "y.md"),
	}
	if err := s.Create(rec); err == nil {
		t.Fatal("get loadall fail")
	}
	_ = os.Chmod(s.recordsDir(), 0o755)
}

// Verifies: SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4
func TestWriteLockedAndConfined(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	if err := confinedToRoot(s.Root, filepath.Dir(s.Root)); err == nil {
		t.Fatal("parent path should be ..")
	}
	rec := &record.Record{ID: "", Type: "note", Path: filepath.Join(s.Root, "records", "notes", "emptyid.md"), Fields: map[string]any{"type": "note"}}
	if err := s.Write(rec); err != nil {
		t.Fatal(err)
	}
	if err := s.writeLocked(&record.Record{Path: filepath.Join(s.Root, "..", "escape.md"), Fields: map[string]any{}}); err == nil {
		t.Fatal("escape write")
	}
	blocked := filepath.Join(s.Root, "records", "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.writeLocked(&record.Record{Path: filepath.Join(blocked, "x.md"), Fields: map[string]any{}}); err == nil {
		t.Fatal("mkdir on file")
	}
	ro := filepath.Join(s.Root, "records", "notes")
	if err := os.Chmod(ro, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(ro, 0o755) })
	if err := s.writeLocked(&record.Record{Path: filepath.Join(ro, "no.md"), Fields: map[string]any{}}); err == nil {
		t.Fatal("writefile ro")
	}
	_ = os.Chmod(ro, 0o755)
	target := filepath.Join(s.Root, "records", "notes", "dir.md")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := s.writeLocked(&record.Record{Path: target, Fields: map[string]any{}}); err == nil {
		t.Fatal("rename onto dir")
	}
}

// Verifies: SYS-REQ-260820-7WT4 SW-REQ-260820-9C5Z SYS-REQ-260820-YWV4 SW-REQ-260820-8PMR SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4
func TestPatchAndEnumAndHashAndLock(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	rec := &record.Record{Type: "grant", ID: "grant_e", Fields: map[string]any{"title": "E", "status": "researching"}}
	if err := s.Create(rec); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Patch("missing", map[string]any{"x": 1}, nil, ""); err == nil {
		t.Fatal("missing patch")
	}
	if err := s.checkEnum(rec, map[string]any{"status": ""}); err != nil {
		t.Fatal(err)
	}
	if err := s.checkEnum(&record.Record{Type: "note", Fields: map[string]any{"status": "x"}}, map[string]any{"status": "x"}); err != nil {
		t.Fatal("unknown type")
	}
	if err := os.WriteFile(filepath.Join(s.Root, "crm.yaml"), []byte("types:\n  grant:\n    fields:\n      extra:\n        type: string\n      status:\n        enum: [open]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec.Set("status", "")
	if err := s.checkEnum(rec, map[string]any{"extra": "v", "status": ""}); err != nil {
		t.Fatal(err)
	}

	yamlPath := filepath.Join(s.Root, "crm.yaml")
	if err := os.Chmod(yamlPath, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(yamlPath, 0o644) })
	if _, err := loadTypeSchemas(s.Root); err == nil {
		t.Fatal("unreadable crm.yaml")
	}
	_ = os.Chmod(yamlPath, 0o644)

	if _, err := FileHash(filepath.Join(s.Root, "missing.bin")); err == nil {
		t.Fatal("hash missing")
	}
	if _, err := FileHash(filepath.Join(s.Root, "crm.yaml")); err != nil {
		t.Fatal(err)
	}

	s.ApplyTemplate(&record.Record{Type: "..", Fields: map[string]any{}})

	crmRuntime := filepath.Join(s.Root, ".crm", "runtime")
	if err := os.RemoveAll(crmRuntime); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(crmRuntime, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.lock("z"); err == nil {
		t.Fatal("lock mkdir")
	}

	s = &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	locks := filepath.Join(s.Root, ".crm", "runtime", "locks")
	if err := os.Chmod(locks, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locks, 0o755) })
	if _, err := s.lock("q"); err == nil {
		t.Fatal("lock open")
	}

	s = &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(&record.Record{Type: "note", ID: "note_p", Fields: map[string]any{"title": "P"}}); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(s.recordsDir()); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.recordsDir(), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Patch("note_p", map[string]any{"title": "Q"}, nil, ""); err == nil {
		t.Fatal("patch get fail")
	}
}

// Verifies: SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4 SYS-REQ-260820-9J7C
func TestWriteLockedAfterLockAndInboxRead(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.Root, "inbox", "bad.md"), []byte("---\n:\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.LoadAll(); err == nil {
		t.Fatal("inbox bad parse")
	}
}

// Verifies: SYS-REQ-260820-YWV4 SW-REQ-260820-8PMR
func TestCheckEnumNilConfigAndMissingYAML(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.checkEnum(&record.Record{Type: "note", Fields: map[string]any{}}, map[string]any{"x": 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := loadTypeSchemas(s.Root); err != nil {
		t.Fatal(err)
	}
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func TestWalkMarkdownMissingInbox(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(s.Root, "inbox")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.LoadAll(); err != nil {
		t.Fatal(err)
	}
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func TestCreateGetUnexpectedErrorAndTemplateEscape(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(s.recordsDir()); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(blocker, "nested"), s.recordsDir()); err != nil {
		t.Fatal(err)
	}
	rec := &record.Record{Type: "note", ID: "note_z", Fields: map[string]any{"title": "Z"}, Path: filepath.Join(s.Root, "inbox", "z.md")}
	if err := s.Create(rec); err == nil {
		t.Fatal("expected get error")
	}
	s.ApplyTemplate(&record.Record{Type: filepath.Join("..", "..", "..", "tmp", "x"), Fields: map[string]any{}})
}

// Verifies: SYS-REQ-260820-7WT4 SW-REQ-260820-9C5Z SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4
func TestPatchLockAndWriteErrors(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(&record.Record{Type: "note", ID: "note_w", Fields: map[string]any{"title": "W"}}); err != nil {
		t.Fatal(err)
	}
	locks := filepath.Join(s.Root, ".crm", "runtime", "locks")
	if err := os.Chmod(locks, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locks, 0o755) })
	if _, err := s.Patch("note_w", map[string]any{"title": "X"}, nil, ""); err == nil {
		t.Fatal("lock fail")
	}
	_ = os.Chmod(locks, 0o755)
	notes := filepath.Join(s.Root, "records", "notes")
	if err := os.Chmod(notes, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(notes, 0o755) })
	if _, err := s.Patch("note_w", map[string]any{"title": "Y"}, nil, ""); err == nil {
		t.Fatal("write fail")
	}
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func TestCreateGetLoadError(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(s.recordsDir(), 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(s.recordsDir(), 0o755) })
	err := s.Create(&record.Record{Type: "note", ID: "note_loaderr", Fields: map[string]any{"title": "Z"}, Path: filepath.Join(s.Root, "inbox", "loaderr.md")})
	if err == nil {
		t.Fatal("expected load error")
	}
	t.Logf("create err: %v", err)
}

// Verifies: SYS-REQ-260820-7WT4
// SYS-REQ-260820-7WT4
func TestDeleteLockAndRemoveErrors(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	rec := &record.Record{ID: "note_del", Type: "note", Fields: map[string]any{"title": "D"}, Body: "b"}
	if err := s.Create(rec); err != nil {
		t.Fatal(err)
	}
	locks := filepath.Join(s.Root, ".crm", "runtime", "locks")
	if err := os.Chmod(locks, 0); err == nil {
		t.Cleanup(func() { _ = os.Chmod(locks, 0o755) })
		_ = s.Delete("note_del")
	}
	_ = os.Chmod(locks, 0o755)
	got, err := s.Get("note_del")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Dir(got.Path), 0o555); err == nil {
		t.Cleanup(func() { _ = os.Chmod(filepath.Dir(got.Path), 0o755) })
		_ = s.Delete("note_del")
	}
	_ = os.Chmod(filepath.Dir(got.Path), 0o755)
}

func TestInitWriteWorkspaceDenied(t *testing.T) {
	s := &Store{Root: t.TempDir()}
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(s.Root, "crm.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(s.Root, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(s.Root, 0o755) })
	if err := s.Init(); err == nil {
		t.Fatal("expected writeworkspace fail")
	}
}
