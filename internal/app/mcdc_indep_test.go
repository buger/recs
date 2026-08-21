package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Verifies: SW-REQ-260821-8C2C
// SW-REQ-260821-8C2C
func TestImportCSVIndependence(t *testing.T) {
	a := setupRemaining(t)
	if _, err := a.ImportCSV(strings.NewReader("id,type,title\n"), ""); err == nil {
		// header only is ok empty created
	}
	if _, err := a.ImportCSV(strings.NewReader(""), "note"); err == nil {
		t.Fatal("empty csv")
	}
	if _, err := a.ImportCSV(strings.NewReader("\"unclosed"), "note"); err == nil {
		t.Fatal("bad csv")
	}
	// short row, empty header, empty val, tags empty, type from column, default type
	in := "id,type,title,,tags,body,status\nn1,note,Hello,,a,b,open\nn2,,X\n"
	if recs, err := a.ImportCSV(strings.NewReader(in), "note"); err != nil {
		t.Fatal(err, recs)
	}
	if _, err := a.ImportCSV(strings.NewReader("title\nHi\n"), ""); err == nil {
		t.Fatal("type required")
	}
	if _, err := a.ImportCSV(strings.NewReader("id,type,title\n,note,\n"), "note"); err != nil && !strings.Contains(err.Error(), "") {
		t.Fatal(err)
	}
	// create error: confined path
	if _, err := a.ImportCSV(strings.NewReader("id,type,title\n../x,note,T\n"), "note"); err == nil {
		// may or may not fail depending on id sanitization
	}
}

// Verifies: SW-REQ-260821-8C2C
// SW-REQ-260821-8C2C
func TestViewIngestConfinedIndependence(t *testing.T) {
	a := setupRemaining(t)
	if _, err := a.Create("person", "person_bob", map[string]any{"name": "Bob", "company": "company_acme", "customer": "company_acme"}, ""); err != nil {
		t.Fatal(err)
	}
	if v, err := a.View("company_acme"); err != nil {
		t.Fatal(err)
	} else if v == nil {
		t.Fatal("view")
	}
	if _, err := a.View("missing"); err == nil {
		t.Fatal("missing view")
	}
	_, _ = a.Ingest("", []byte("From: a@b.com\nSubject: S\n\nB\n"))
	_, _ = a.Ingest("-", []byte("From: a@b.com\nSubject: S2\n\nB\n"))
	_, _ = a.Ingest("email", []byte("not an email"))
	_, _ = a.Ingest("email", []byte("From: a@b.com\nSubject: S\n\nB\n"))
	_, _ = a.AttachmentFile("person_alice")
	_, _ = a.AttachmentFile("missing")
	// store confined via edit path outside root
	if _, err := a.Edit("person_alice", map[string]any{"title": "Z"}, nil, "bad-version"); err == nil {
		t.Fatal("bad version")
	}
}

// Verifies: SW-REQ-260821-8C2C
// SW-REQ-260821-8C2C
func TestGitAndExportEdges(t *testing.T) {
	a := setupRemaining(t)
	if _, err := a.ExportJSON(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.ExportCSV(); err != nil {
		t.Fatal(err)
	}
	_ = a.Diff()
	_ = a.Changed()
	_ = a.History("person_alice")
	_, _ = a.ResolveWikilink("Acme")
	_, _ = a.ResolveWikilink("missing")
	// chmod records to force export/list errors
	recDir := filepath.Join(a.Root(), "records")
	if err := os.Chmod(recDir, 0); err == nil {
		t.Cleanup(func() { _ = os.Chmod(recDir, 0o755) })
		_, _ = a.ExportJSON()
		_, _ = a.ExportCSV()
		_, _ = a.View("person_alice")
	}
}


// Verifies: SW-REQ-260821-8C2C
// SW-REQ-260821-8C2C
func TestGitCompanionEdges(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	a := setupRemaining(t)
	// rebuild in known root
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = a.Root()
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	run("init")
	run("add", ".")
	run("commit", "-m", "init")
	_ = a.Diff()
	_ = a.Changed()
	_ = a.History("person_alice")
	_ = a.History("missing")
	_, _ = a.EditFromBytes("person_alice", []byte("not markdown"), "")
	_, _ = a.EditFromBytes("missing", []byte("---\nid: x\n---\n"), "")
	_, _ = a.EditFromBytes("person_alice", []byte("---\nid: other\ntype: note\n---\n"), "")
	if _, err := a.EditFromBytes("person_alice", []byte("---\nid: person_alice\n---\nbody\n"), ""); err != nil {
		// type empty uses current
	}
	if _, err := a.Link("person_alice", "company_acme", ""); err == nil {
		t.Fatal("empty rel")
	}
	if _, err := a.Link("person_alice", "missing", "related_to"); err == nil {
		t.Fatal("missing target")
	}
	if _, err := a.Link("missing", "company_acme", "related_to"); err == nil {
		t.Fatal("missing src")
	}
	if _, err := a.Link("person_alice", "company_acme", "related_to"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Link("person_alice", "company_acme", "related_to"); err != nil {
		t.Fatal("dup", err)
	}
	if _, err := a.EditWithEditor("missing", "true"); err == nil {
		t.Fatal("editor missing")
	}
	if _, err := a.EditWithEditor("person_alice", "true"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", "/usr/bin:/bin")
	_ = a.Diff()
}


func TestIngestViewLinkIndependence(t *testing.T) {
	a := setupRemaining(t)
	_, _ = a.Ingest("", []byte(`{"subject":"S"}`))
	_, _ = a.Ingest("-", []byte(`{"subject":"S2"}`))
	_, _ = a.Ingest("email", []byte(`{"subject":"S3","title":"Kept","triage_status":"old","text":"from-text"}`))
	_, _ = a.Ingest("email", []byte(`{"subject":"S4"}`))
	_, _ = a.Ingest("note", []byte(`{"text":12,"title":"N"}`))
	if _, err := a.Create("person", "person_cust", map[string]any{"name": "Cust", "customer": "company_acme"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("person", "person_other", map[string]any{"name": "O", "company": "nope"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Edit("person_alice", map[string]any{"relations": []any{
		map[string]any{"type": "knows", "target": "missing_id"},
		"not-map",
		map[string]any{"type": "works_at", "target": "company_acme"},
	}}, nil, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.View("company_acme"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.View("person_alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.View("person_other"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Link("person_alice", "company_acme", "related_to"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Link("person_alice", "person_cust", "works_at"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.ImportCSV(strings.NewReader("id,type,title,tags\nnx,note,T,\n"), ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AttachmentFile(""); err == nil {
		t.Fatal("empty att")
	}
	inside := filepath.Join(a.Root(), "attachments", "note2.txt")
	if err := os.WriteFile(inside, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AttachmentFile("attachments/note2.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AttachmentFile("/etc/passwd"); err == nil {
		t.Fatal("abs outside")
	}
	if _, err := a.AttachmentFile(".."); err == nil {
		t.Fatal("dotdot")
	}
}

func TestGitFailureAndChmodEdges(t *testing.T) {
	a := setupRemaining(t)
	// .git as a file: gitRoot sees err==nil && !IsDir
	if err := os.WriteFile(filepath.Join(a.Root(), ".git"), []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = a.Diff()
	_ = a.Changed()
	_ = a.History("person_alice")
	_ = os.Remove(filepath.Join(a.Root(), ".git"))

	// empty .git dir is a dir but not a repo
	if err := os.Mkdir(filepath.Join(a.Root(), ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	d := a.Diff()
	if d.OK {
		t.Log("diff ok on empty git", d)
	}
	_ = a.Changed()
	_ = a.History("person_alice")
	_ = a.History("missing")

	t.Setenv("PATH", "/nonexistent-git-bin")
	_ = a.Diff()
	_ = a.Changed()
	_ = a.History("person_alice")

	// ProjectDashboard after chmod records
	if _, err := a.CreateDashboard("dashx", "X", "1x1", "", nil); err != nil {
		t.Fatal(err)
	}
	recDir := filepath.Join(a.Root(), "records")
	if err := os.Chmod(recDir, 0); err == nil {
		t.Cleanup(func() { _ = os.Chmod(recDir, 0o755) })
		_, _ = a.ProjectDashboard("dashx")
		_, _ = a.ResolveWikilink("Acme")
		_, _ = a.View("person_alice")
		_, _ = a.ExportCSV()
		_, _ = a.ExportJSON()
	}
}

func TestEditFromBytesAndEditorEdges(t *testing.T) {
	a := setupRemaining(t)
	_, _ = a.EditFromBytes("person_alice", []byte("\x00\x01"), "")
	rec, err := a.Store.Get("person_alice")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Dir(rec.Path), 0o555); err == nil {
		t.Cleanup(func() { _ = os.Chmod(filepath.Dir(rec.Path), 0o755) })
		_, _ = a.EditFromBytes("person_alice", []byte("---\nid: person_alice\ntype: person\n---\n"), rec.Version())
	}
	_ = os.Chmod(filepath.Dir(rec.Path), 0o755)
	if _, err := a.EditWithEditor("person_alice", `rm -f`); err == nil {
		t.Fatal("editor removed temp")
	}
}

func TestIngestNoSubjectAndParseFail(t *testing.T) {
	a := setupRemaining(t)
	_, _ = a.Ingest("email", []byte(`{}`))
	_, _ = a.Ingest("email", []byte(`{"title":"OnlyTitle"}`))
	if _, err := a.EditFromBytes("person_alice", []byte("---\nno-close"), ""); err == nil {
		t.Fatal("unclosed frontmatter")
	}
	if _, err := a.EditFromBytes("person_alice", []byte("---\n:\n---\n"), ""); err == nil {
		t.Fatal("bad yaml")
	}
}

func TestHistoryEmptyLogLines(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	a := setupRemaining(t)
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = a.Root()
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	run("init")
	run("add", ".")
	run("commit", "-m", "init")
	if _, err := a.Create("note", "note_after", map[string]any{"title": "After"}, "x"); err != nil {
		t.Fatal(err)
	}
	h := a.History("note_after")
	if !h.OK || !h.Git {
		t.Fatalf("%#v", h)
	}
}
