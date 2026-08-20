package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"crm/internal/cli"
)

func TestCLIValidateFailEnumConflictAndStoreErrors(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	must := func(code int, what string) {
		t.Helper()
		if code != 0 {
			t.Fatalf("%s: code=%d out=%s err=%s", what, code, out.String(), errb.String())
		}
	}
	must(cli.Main([]string{"--root", root, "init"}, out, errb), "init")
	must(cli.Main([]string{"--root", root, "create", "note", "--id", "note_v", "--title", "V"}, out, errb), "create")

	if err := os.WriteFile(filepath.Join(root, "crm.yaml"), []byte("types:\n  note:\n    required: [owner]\n    fields:\n      status:\n        enum: [open]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := cli.Main([]string{"--root", root, "validate"}, out, errb); code != 1 {
		t.Fatalf("validate want 1 got %d out=%s err=%s", code, out.String(), errb.String())
	}
	if cli.Main([]string{"--root", root, "patch", "note_v", "--set", "status=nope"}, out, errb) == 0 {
		t.Fatal("enum human")
	}
	if cli.Main([]string{"--root", root, "patch", "note_v", "--if-version", "sha256:dead", "--set", "title=Z"}, out, errb) == 0 {
		t.Fatal("conflict human")
	}
	if cli.Main([]string{"--root", root, "--md", "--json", "context", "note_v"}, out, errb) != 0 {
		t.Fatal("md+json", errb.String())
	}
	if cli.Main([]string{"--root", root, "--md", "context", "note_v"}, out, errb) != 0 {
		t.Fatal("md", errb.String())
	}

	file := filepath.Join(t.TempDir(), "notdir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if cli.Main([]string{"--root", file, "init"}, out, errb) == 0 {
		t.Fatal("init on file")
	}

	records := filepath.Join(root, "records")
	if err := os.Chmod(records, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(records, 0o755) })
	for _, args := range [][]string{
		{"list"},
		{"search", "x"},
		{"query", "status = open"},
		{"board"},
		{"next"},
		{"triage"},
		{"validate"},
		{"index"},
		{"inbox"},
		{"context", "note_v"},
		{"move", "note_v", "grants", "open"},
		{"agent", "install"},
		{"set", "note_v", "title", "Z"},
		{"show", "note_v"},
	} {
		all := append([]string{"--root", root}, args...)
		code := cli.Main(all, out, errb)
		if args[0] == "board" || args[0] == "agent" {
			continue
		}
		if code == 0 {
			t.Fatalf("expected error for %v out=%s err=%s", args, out.String(), errb.String())
		}
	}
}

func TestCLIAgentInstallBlocked(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	_ = os.Remove(filepath.Join(root, "AGENTS.md"))
	if err := os.Mkdir(filepath.Join(root, "AGENTS.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	if cli.Main([]string{"--root", root, "agent", "install"}, out, errb) == 0 {
		t.Fatal("agent write to dir")
	}
}

func TestCLIBoardListError(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	boards := filepath.Join(root, "boards")
	if err := os.RemoveAll(boards); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(boards, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if cli.Main([]string{"--root", root, "board"}, out, errb) == 0 {
		t.Fatal("expected board list error")
	}
}
