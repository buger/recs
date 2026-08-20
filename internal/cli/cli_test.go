package cli_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crm/internal/app"
	"crm/internal/cli"
	"crm/internal/serve"
)

// Verifies: SYS-REQ-260820-PG9C SW-REQ-260820-YB5C INT-REQ-260820-JC9M SYS-REQ-260820-5C9D SW-REQ-260820-ZKCV SYS-REQ-260820-DCG4 SW-REQ-260820-D5WE
func TestCLIJSONAndTriage(t *testing.T) {
	root := t.TempDir()
	if code := run(t, root, "init", "--json"); code != 0 {
		t.Fatal(code)
	}
	if code := run(t, root, "create", "grant", "--id", "grant_x", "--title", "X", "--set", "status=inbox"); code != 0 {
		t.Fatal(code)
	}
	out := &bytes.Buffer{}
	if code := cli.Main([]string{"--root", root, "list", "--json"}, out, out); code != 0 {
		t.Fatal(out.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil || payload["ok"] != true {
		t.Fatalf("%s", out.String())
	}
	out.Reset()
	if code := cli.Main([]string{"--root", root, "triage", "--json"}, out, out); code != 0 {
		t.Fatal(out.String())
	}
	if !strings.Contains(out.String(), "inbox") {
		t.Fatal(out.String())
	}
	out.Reset()
	if code := cli.Main([]string{"--root", root, "next", "--json"}, out, out); code != 0 {
		t.Fatal(out.String())
	}
}

// Verifies: INT-REQ-260820-AHKR SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7
func TestServeUsesAppLayer(t *testing.T) {
	root := t.TempDir()
	a := app.OpenOrCWD(root)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("grant", "grant_web", map[string]any{"title": "Web", "status": "researching"}, ""); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(serve.Handler(a))
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/api/records")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true {
		t.Fatalf("%v", payload)
	}
	body, _ := json.Marshal(map[string]string{"id": "grant_web", "column": "applied"})
	resp, err = http.Post(ts.URL+"/api/boards/grants/move", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	got, err := a.Show("grant_web")
	if err != nil || got.GetString("status") != "applied" {
		t.Fatalf("http move did not use app layer: %v %#v", err, got)
	}
}

func run(t *testing.T, root string, args ...string) int {
	t.Helper()
	var out bytes.Buffer
	all := append([]string{"--root", root}, args...)
	code := cli.Main(all, &out, &out)
	if code != 0 {
		t.Log(out.String())
	}
	_ = filepath.Separator
	return code
}

// Verifies: SYS-REQ-260820-PG9C SW-REQ-260820-YB5C SYS-REQ-260820-DCG4 SW-REQ-260820-D5WE SYS-REQ-260820-9J7C SW-REQ-260820-N02Y SYS-REQ-260820-2SQZ
func TestAgentInboxTemplateAndEnum(t *testing.T) {
	root := t.TempDir()
	if code := run(t, root, "init"); code != 0 {
		t.Fatal(code)
	}
	if code := run(t, root, "agent", "install", "--json"); code != 0 {
		t.Fatal("agent install")
	}
	for _, name := range []string{"AGENTS.md", "SKILL.md"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatalf("missing %s", name)
		}
	}
	if code := run(t, root, "create", "grant", "--id", "grant_tpl", "--title", "Template Grant"); code != 0 {
		t.Fatal("create")
	}
	recOut := &bytes.Buffer{}
	if code := cli.Main([]string{"--root", root, "show", "grant_tpl"}, recOut, recOut); code != 0 {
		t.Fatal(recOut.String())
	}
	if !strings.Contains(recOut.String(), "## Opportunity") {
		t.Fatalf("template body missing: %s", recOut.String())
	}
	inbox := filepath.Join(root, "inbox", "note.md")
	if err := os.WriteFile(inbox, []byte("---\nid: note_in\ntype: note\nstatus: inbox\n---\n\n# In\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := &bytes.Buffer{}
	if code := cli.Main([]string{"--root", root, "inbox", "--json"}, out, out); code != 0 {
		t.Fatal(out.String())
	}
	if !strings.Contains(out.String(), "note_in") {
		t.Fatalf("inbox missing file: %s", out.String())
	}
	out.Reset()
	if code := cli.Main([]string{"--root", root, "set", "grant_tpl", "status", "nope", "--json"}, out, out); code != 1 {
		t.Fatalf("expected enum fail: %s", out.String())
	}
	if !strings.Contains(out.String(), "invalid_enum") || !strings.Contains(out.String(), "allowed") {
		t.Fatalf("structured enum: %s", out.String())
	}
}

func TestConflictJSONAndDeadlineTriage(t *testing.T) {
	root := t.TempDir()
	if code := run(t, root, "init"); code != 0 {
		t.Fatal(code)
	}
	if code := run(t, root, "create", "grant", "--id", "grant_c", "--title", "C", "--set", "status=researching"); code != 0 {
		t.Fatal(code)
	}
	out := &bytes.Buffer{}
	if code := cli.Main([]string{"--root", root, "patch", "grant_c", "--set", "status=applied", "--if-version", "sha256:dead", "--json"}, out, out); code != 1 {
		t.Fatalf("expected conflict: %s", out.String())
	}
	if !strings.Contains(out.String(), "conflict") || !strings.Contains(out.String(), "expected_version") {
		t.Fatalf("structured conflict: %s", out.String())
	}
	if code := run(t, root, "create", "grant", "--id", "grant_due", "--title", "Due", "--set", "status=preparing", "--set", "deadline=2000-01-01"); code != 0 {
		t.Fatal("create due")
	}
	out.Reset()
	if code := cli.Main([]string{"--root", root, "triage", "--json"}, out, out); code != 0 {
		t.Fatal(out.String())
	}
	if !strings.Contains(out.String(), "grant_due") || !strings.Contains(out.String(), "overdue") {
		t.Fatalf("deadline triage: %s", out.String())
	}
}
