package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crm/internal/cli"
)

// Verifies: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C INT-REQ-260821-8HAC SYS-REQ-260821-QF1J
// SYS-REQ-260821-JYEJ:nominal:nominal
// SYS-REQ-260821-JYEJ:error_handling:negative
// SW-REQ-260821-8C2C:nominal:nominal
// INT-REQ-260821-8HAC:nominal:nominal
func TestRemainingCLICommands(t *testing.T) {
	root := t.TempDir()
	if run(t, root, "init") != 0 {
		t.Fatal("init")
	}
	if run(t, root, "create", "person", "--id", "person_alice", "--name", "Alice Smith") != 0 {
		t.Fatal("alice")
	}
	if run(t, root, "create", "company", "--id", "company_acme", "--name", "Acme") != 0 {
		t.Fatal("acme")
	}
	if run(t, root, "edit", "person_alice", "--set", "status=active", "--body", "hello") != 0 {
		t.Fatal("edit")
	}
	if run(t, root, "edit") == 0 {
		t.Fatal("edit usage")
	}
	if run(t, root, "link", "person_alice", "company_acme", "--relation", "works_at", "--json") != 0 {
		t.Fatal("link")
	}
	if run(t, root, "link", "person_alice") == 0 {
		t.Fatal("link usage")
	}
	out := &bytes.Buffer{}
	cli.Input = strings.NewReader(`{"subject":"Need help","from":{"email":"a@x.com"},"body":"blocked"}`)
	if cli.Main([]string{"--root", root, "ingest", "email", "--json"}, out, out) != 0 {
		t.Fatal(out.String())
	}
	if !strings.Contains(out.String(), "email") {
		t.Fatal(out.String())
	}
	tmp := filepath.Join(root, "mail.json")
	os.WriteFile(tmp, []byte(`{"type":"email","subject":"File","body":"x"}`), 0o644)
	if run(t, root, "ingest", tmp, "--json") != 0 {
		t.Fatal("ingest file")
	}
	if run(t, root, "export", "--json") != 0 {
		t.Fatal("export json")
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "export", "--csv"}, out, out) != 0 || !strings.Contains(out.String(), "person_alice") {
		t.Fatal(out.String())
	}
	csvPath := filepath.Join(root, "in.csv")
	os.WriteFile(csvPath, []byte("type,title,status\nnote,FromCSV,open\n"), 0o644)
	if run(t, root, "import", csvPath, "--json") != 0 {
		t.Fatal("import")
	}
	if run(t, root, "import") == 0 {
		t.Fatal("import usage")
	}
	if run(t, root, "diff", "--json") != 0 || run(t, root, "changed", "--json") != 0 {
		t.Fatal("git empty")
	}
	if run(t, root, "history", "person_alice", "--json") != 0 {
		t.Fatal("history")
	}
	if run(t, root, "history") == 0 {
		t.Fatal("history usage")
	}
	if run(t, root, "delete", "person_alice", "--json") != 0 {
		t.Fatal("delete")
	}
	if run(t, root, "delete", "person_alice", "--json") == 0 {
		t.Fatal("delete missing")
	}
	if run(t, root, "delete") == 0 {
		t.Fatal("delete usage")
	}
	help := &bytes.Buffer{}
	cli.Main([]string{"--help"}, help, help)
	for _, cmd := range []string{"edit", "delete", "link", "ingest", "export", "import", "diff", "changed", "history"} {
		if !strings.Contains(help.String(), cmd) {
			t.Fatalf("help missing %s", cmd)
		}
	}
	var payload map[string]any
	out.Reset()
	cli.Main([]string{"--root", root, "export", "--csv", "--json"}, out, out)
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil || payload["ok"] != true {
		t.Fatal(out.String())
	}
}

// Verifies: SYS-REQ-260821-JYEJ
func TestEditEditorAndIngestDash(t *testing.T) {
	root := t.TempDir()
	if run(t, root, "init") != 0 {
		t.Fatal("init")
	}
	if run(t, root, "create", "note", "--id", "note_x", "--title", "X") != 0 {
		t.Fatal("create")
	}
	t.Setenv("EDITOR", "true")
	if run(t, root, "edit", "note_x") != 0 {
		t.Fatal("editor")
	}
	t.Setenv("EDITOR", "")
	if run(t, root, "edit", "note_x") == 0 {
		t.Fatal("no editor")
	}
	out := &bytes.Buffer{}
	cli.Input = strings.NewReader(`{"type":"note","title":"stdin"}`)
	if cli.Main([]string{"--root", root, "ingest", "-"}, out, out) != 0 {
		t.Fatal(out.String())
	}
}
