package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/buger/recs/internal/app"
)

// Verifies: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func TestGitCompanionWithRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	a := app.OpenOrCWD(root)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("note", "note_g", map[string]any{"title": "G"}, "body"); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	run("init")
	run("add", "crm.yaml", "records")
	run("commit", "-m", "init")
	if _, err := a.Edit("note_g", map[string]any{"status": "open"}, nil, ""); err != nil {
		t.Fatal(err)
	}
	d := a.Diff()
	if !d.OK || !d.Git {
		t.Fatalf("%#v", d)
	}
	c := a.Changed()
	if !c.OK || !c.Git {
		t.Fatalf("%#v", c)
	}
	h := a.History("note_g")
	if !h.OK || !h.Git || len(h.History) == 0 {
		t.Fatalf("%#v", h)
	}
	_ = filepath.Separator
}
