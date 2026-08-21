package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crm/internal/app"
	"crm/internal/dist"
)

func setupRemaining(t *testing.T) *app.App {
	t.Helper()
	a := app.OpenOrCWD(t.TempDir())
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("person", "person_alice", map[string]any{"name": "Alice Smith"}, "Talked with [[Acme]].\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("company", "company_acme", map[string]any{"name": "Acme"}, "Acme notes\n"); err != nil {
		t.Fatal(err)
	}
	return a
}

// Verifies: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C INT-REQ-260821-8HAC STK-REQ-260821-QTPP
// SYS-REQ-260821-JYEJ:nominal:nominal
// SYS-REQ-260821-JYEJ:error_handling:negative
// SW-REQ-260821-8C2C:nominal:nominal
// SW-REQ-260821-8C2C:error_handling:negative
// INT-REQ-260821-8HAC:nominal:nominal
// INT-REQ-260821-8HAC:integration:integration
// STK-REQ-260821-QTPP:nominal:nominal
func TestEditDeleteLinkIngestExportImport(t *testing.T) {
	a := setupRemaining(t)
	body := "new body"
	res, err := a.Edit("person_alice", map[string]any{"status": "active"}, &body, "")
	if err != nil || res.Record.Body != "new body" || res.Record.GetString("status") != "active" {
		t.Fatalf("edit %v %#v", err, res)
	}
	if _, err := a.Edit("person_alice", nil, nil, ""); err == nil {
		t.Fatal("empty edit")
	}
	if _, err := a.Link("person_alice", "company_acme", "works_at"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Link("person_alice", "company_acme", "works_at"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Link("person_alice", "missing", "knows"); err == nil {
		t.Fatal("missing target")
	}
	view, err := a.View("person_alice")
	if err != nil || len(view.Relations) == 0 || view.HTML == "" {
		t.Fatalf("view %v %#v", err, view)
	}
	back, err := a.View("company_acme")
	if err != nil || len(back.Backlinks) == 0 {
		t.Fatalf("backlinks %#v", back)
	}
	attDir := filepath.Join(a.Root(), "attachments")
	os.WriteFile(filepath.Join(attDir, "note.txt"), []byte("hi"), 0o644)
	if _, err := a.Edit("person_alice", map[string]any{"attachments": []any{"attachments/note.txt"}}, nil, ""); err != nil {
		t.Fatal(err)
	}
	view, _ = a.View("person_alice")
	if len(view.Attachments) != 1 {
		t.Fatal(view.Attachments)
	}
	if p, err := a.AttachmentFile("attachments/note.txt"); err != nil || p == "" {
		t.Fatal(err)
	}
	if _, err := a.AttachmentFile("../etc/passwd"); err == nil {
		t.Fatal("escape")
	}
	raw := []byte(`{"subject":"Hello","from":{"email":"a@x.com"},"body":"hi"}`)
	ing, err := a.Ingest("email", raw)
	if err != nil || ing.Type != "email" || ing.GetString("title") != "Hello" {
		t.Fatalf("ingest %v %#v", err, ing)
	}
	if _, err := a.Ingest("", []byte(`{`)); err == nil {
		t.Fatal("bad json")
	}
	js, err := a.ExportJSON()
	if err != nil || len(js) < 3 {
		t.Fatal(js, err)
	}
	csv, err := a.ExportCSV()
	if err != nil || !strings.Contains(csv, "person_alice") {
		t.Fatal(csv, err)
	}
	imp := strings.NewReader("type,title,status\nnote,Imported,open\n")
	got, err := a.ImportCSV(imp, "")
	if err != nil || len(got) != 1 || got[0].Type != "note" {
		t.Fatalf("import %v %#v", err, got)
	}
	imp2 := strings.NewReader("title,status\nNoType,open\n")
	if _, err := a.ImportCSV(imp2, "task"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.ImportCSV(strings.NewReader(""), ""); err == nil {
		t.Fatal("empty csv")
	}
	if err := a.Delete("person_alice"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Show("person_alice"); err == nil {
		t.Fatal("deleted still there")
	}
	if err := a.Delete("person_alice"); err == nil {
		t.Fatal("double delete")
	}
}

// Verifies: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func TestGitCompanionWithoutRepo(t *testing.T) {
	a := setupRemaining(t)
	d := a.Diff()
	if !d.OK || d.Git {
		t.Fatalf("%#v", d)
	}
	c := a.Changed()
	if !c.OK || c.Git || len(c.Changed) != 0 {
		t.Fatalf("%#v", c)
	}
	h := a.History("company_acme")
	if !h.OK || h.Git {
		t.Fatalf("%#v", h)
	}
}

// Verifies: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func TestEditFromBytesAndEditor(t *testing.T) {
	a := setupRemaining(t)
	rec, _ := a.Show("company_acme")
	data := []byte("---\nid: company_acme\ntype: company\nname: Acme\n---\nEdited\n")
	res, err := a.EditFromBytes("company_acme", data, rec.Version())
	if err != nil || !strings.Contains(res.Record.Body, "Edited") {
		t.Fatal(err, res)
	}
	if _, err := a.EditFromBytes("company_acme", data, "sha256:dead"); err == nil {
		t.Fatal("conflict")
	}
	if _, err := a.EditFromBytes("company_acme", []byte("---\nid: other\ntype: company\n---\n"), ""); err == nil {
		t.Fatal("id change")
	}
	t.Setenv("EDITOR", "true")
	if _, err := a.EditWithEditor("company_acme", "true"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.EditWithEditor("company_acme", ""); err == nil {
		t.Fatal("empty editor")
	}
}

// Verifies: SYS-REQ-260821-QF1J
func TestResolveWikilink(t *testing.T) {
	a := setupRemaining(t)
	id, err := a.ResolveWikilink("Alice Smith")
	if err != nil || id != "person_alice" {
		t.Fatal(id, err)
	}
	if _, err := a.ResolveWikilink("nope"); err == nil {
		t.Fatal("missing")
	}
}

// Verifies: SYS-REQ-260821-AFPN SW-REQ-260821-AC3S STK-REQ-260820-T8AZ
func TestAPEPackagingFilesExist(t *testing.T) {
	if dist.BuildAPE() != "scripts/build-ape.sh" {
		t.Fatal(dist.BuildAPE())
	}
	dir, _ := os.Getwd()
	root := ""
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			root = dir
			break
		}
		dir = filepath.Dir(dir)
	}
	if root == "" {
		t.Fatal("go.mod not found")
	}
	for _, rel := range []string{"scripts/build-ape.sh", "ape/wrapper.c", "ape/hello.c", "Makefile", "docs/distribution.md"} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatal(rel, err)
		}
	}
}

