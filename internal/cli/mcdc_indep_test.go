package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buger/recs/internal/cli"
)

// Verifies: SW-REQ-260821-8C2C
// SW-REQ-260821-8C2C
func TestCLIMainIndependence(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	must := func(code int, what string) {
		t.Helper()
		if code != 0 {
			t.Fatalf("%s: %d out=%s err=%s", what, code, out.String(), errb.String())
		}
	}
	must(cli.Main([]string{"--root", root, "init"}, out, errb), "init")
	must(cli.Main([]string{"--root", root, "create", "note", "--id", "note_a", "--title", "A", "--body", "hello"}, out, errb), "create")
	must(cli.Main([]string{"--root", root, "--relation=related_to", "link", "note_a", "note_a"}, out, errb), "link eq")
	if cli.Main([]string{"--root", root, "link", "note_a", "note_a"}, out, errb) == 0 {
		t.Fatal("link without relation")
	}
	if cli.Main([]string{"--root", root, "link", "note_a"}, out, errb) == 0 {
		t.Fatal("link short")
	}
	must(cli.Main([]string{"--root", root, "edit", "note_a", "--set", "status=open"}, out, errb), "edit set")
	must(cli.Main([]string{"--root", root, "edit", "note_a", "--body", "only-body"}, out, errb), "edit body")
	t.Setenv("EDITOR", "")
	if cli.Main([]string{"--root", root, "edit", "note_a"}, out, errb) == 0 {
		t.Fatal("edit needs editor")
	}
	if cli.Main([]string{"--root", root, "edit", "missing", "--set", "status=open"}, out, errb) == 0 {
		t.Fatal("edit missing")
	}
	t.Setenv("EDITOR", "false")
	if cli.Main([]string{"--root", root, "edit", "note_a"}, out, errb) == 0 {
		t.Fatal("bad editor")
	}

	must(cli.Main([]string{"--root", root, "dashboard", "new", "dash_a", "--name", "D"}, out, errb), "dash no type")
	must(cli.Main([]string{"--root", root, "dashboard", "new", "dash_b", "--name", "D", "--set", "type=count", "--set", "query=status=open"}, out, errb), "dash type")
	if cli.Main([]string{"--root", root, "dashboard", "new"}, out, errb) == 0 {
		t.Fatal("dash new usage")
	}
	if cli.Main([]string{"--root", root, "dashboard", "missing"}, out, errb) == 0 {
		t.Fatal("dash missing")
	}
	must(cli.Main([]string{"--root", root, "dashboard"}, out, errb), "dash list")
	must(cli.Main([]string{"--root", root, "dashboard", "dash_a"}, out, errb), "dash show")

	if cli.Main([]string{"--root", root, "--json", "not-a-cmd"}, out, errb) == 0 {
		t.Fatal("unknown json")
	}
	if cli.Main([]string{"--root", root, "not-a-cmd"}, out, errb) == 0 {
		t.Fatal("unknown human")
	}

	cli.Input = strings.NewReader("From: a@b.com\nSubject: Hi\n\nBody\n")
	_ = cli.Main([]string{"--root", root, "ingest", "email", "-"}, out, errb)
	cli.Input = strings.NewReader("---\nid: note_in\ntype: note\ntitle: In\n---\nX\n")
	_ = cli.Main([]string{"--root", root, "ingest", "record", "-"}, out, errb)
	src := filepath.Join(root, "ing.md")
	if err := os.WriteFile(src, []byte("---\ntype: note\ntitle: F\n---\nZ\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = cli.Main([]string{"--root", root, "ingest", "record", src}, out, errb)
	if cli.Main([]string{"--root", root, "ingest", "record", filepath.Join(root, "nope.md")}, out, errb) == 0 {
		t.Fatal("ingest missing file")
	}

	csv := filepath.Join(root, "in.csv")
	if err := os.WriteFile(csv, []byte("id,type,title\nn1,note,T\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	must(cli.Main([]string{"--root", root, "import", csv}, out, errb), "import")
	if cli.Main([]string{"--root", root, "import", filepath.Join(root, "no.csv")}, out, errb) == 0 {
		t.Fatal("import missing")
	}
	must(cli.Main([]string{"--root", root, "export"}, out, errb), "export json")
	must(cli.Main([]string{"--root", root, "--csv", "export"}, out, errb), "export csv")
	must(cli.Main([]string{"--root", root, "--csv", "--json", "export"}, out, errb), "export csv json")

	_ = cli.Main([]string{"--root", root, "diff"}, out, errb)
	_ = cli.Main([]string{"--root", root, "changed"}, out, errb)
	_ = cli.Main([]string{"--root", root, "history", "note_a"}, out, errb)

	if cli.Main([]string{"--root", root, "dashboard", "new", "bad id"}, out, errb) == 0 {
		t.Fatal("bad dash id")
	}

	// ListDashboards error
	dashDir := filepath.Join(root, "dashboards")
	if err := os.Chmod(dashDir, 0); err == nil {
		t.Cleanup(func() { _ = os.Chmod(dashDir, 0o755) })
		if cli.Main([]string{"--root", root, "dashboard"}, out, errb) == 0 {
			t.Fatal("dash list unreadable")
		}
	}

	if cli.Main([]string{"--root", root, "create"}, out, errb) == 0 {
		t.Fatal("create usage")
	}
}

// Verifies: SYS-REQ-260821-8FKR
// SYS-REQ-260821-8FKR
func TestCLIGitHumanAndHelpHints(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	if cli.Main([]string{"help"}, out, errb) != 0 {
		t.Fatal("help")
	}
	if cli.Main([]string{"--help"}, out, errb) != 0 {
		t.Fatal("--help")
	}
	if cli.Main([]string{"create", "--help"}, out, errb) != 0 {
		t.Fatal("create help")
	}
	if cli.Main([]string{"--json"}, out, errb) != 0 {
		t.Fatal("json no cmd")
	}
}

func TestCLIMainLeftoverIndependence(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init", errb.String())
	}
	if cli.Main([]string{"--root", root, "create", "note", "--id", "note_a", "--title", "A", "--body", "b"}, out, errb) != 0 {
		t.Fatal("create")
	}
	// empty type: ok && typ==""
	_ = cli.Main([]string{"--root", root, "dashboard", "new", "dash_empty", "--set", "type="}, out, errb)
	// --relation without a following value
	_ = cli.Main([]string{"--root", root, "link", "note_a", "note_a", "--relation"}, out, errb)
	// ingest with no rest (reads stdin)
	cli.Input = strings.NewReader(`{"type":"note","title":"from-stdin"}`)
	_ = cli.Main([]string{"--root", root, "ingest", "-"}, out, errb)
	// help serve has empty JSON shape
	if cli.Main([]string{"help", "serve"}, out, errb) != 0 {
		t.Fatal("help serve")
	}
	// link error path
	if cli.Main([]string{"--root", root, "link", "missing", "note_a", "--relation", "x"}, out, errb) == 0 {
		t.Fatal("link missing")
	}
	// import parse error
	bad := filepath.Join(root, "bad.csv")
	if err := os.WriteFile(bad, []byte("\"unclosed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if cli.Main([]string{"--root", root, "import", bad}, out, errb) == 0 {
		t.Fatal("bad import")
	}
	// export errors
	recDir := filepath.Join(root, "records")
	if err := os.Chmod(recDir, 0); err == nil {
		t.Cleanup(func() { _ = os.Chmod(recDir, 0o755) })
		_ = cli.Main([]string{"--root", root, "--csv", "export"}, out, errb)
		_ = cli.Main([]string{"--root", root, "export"}, out, errb)
	}
	_ = os.Chmod(recDir, 0o755)
	// git commands fail when .git is an empty directory
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	_ = cli.Main([]string{"--root", root, "diff"}, out, errb)
	_ = cli.Main([]string{"--root", root, "changed"}, out, errb)
	_ = cli.Main([]string{"--root", root, "history", "note_a"}, out, errb)
}
