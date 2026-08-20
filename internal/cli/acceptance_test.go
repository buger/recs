package cli_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"crm/internal/app"
	"crm/internal/cli"
	"crm/internal/serve"
)

// STK-REQ-260820-KAGT:AC-001:acceptance
func TestAcceptanceFileRemainsCanonical(t *testing.T) {
	root := t.TempDir()
	if code := run(t, root, "init"); code != 0 {
		t.Fatal("init")
	}
	if code := run(t, root, "create", "grant", "--id", "grant_file", "--title", "File Grant", "--set", "status=researching"); code != 0 {
		t.Fatal("create")
	}
	path := filepath.Join(root, "records", "grants", "grant_file.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "status: researching") || !strings.Contains(text, "File Grant") {
		t.Fatalf("canonical file missing fields: %s", text)
	}
	text = strings.Replace(text, "status: researching", "status: preparing", 1)
	text += "\nEdited by hand.\n"
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	out := &bytes.Buffer{}
	if code := cli.Main([]string{"--root", root, "show", "grant_file", "--json"}, out, out); code != 0 {
		t.Fatal(out.String())
	}
	if !strings.Contains(out.String(), "preparing") || !strings.Contains(out.String(), "Edited by hand") {
		t.Fatalf("hand edit not visible: %s", out.String())
	}
}

// STK-REQ-260820-V5ZD:AC-001:acceptance
func TestAcceptanceJSONHasOK(t *testing.T) {
	root := t.TempDir()
	if code := run(t, root, "init", "--json"); code != 0 {
		t.Fatal("init")
	}
	out := &bytes.Buffer{}
	if code := cli.Main([]string{"--root", root, "create", "grant", "--id", "grant_json", "--title", "JSON", "--json"}, out, out); code != 0 {
		t.Fatal(out.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != true {
		t.Fatalf("ok missing: %s", out.String())
	}
	out.Reset()
	if code := cli.Main([]string{"--root", root, "list", "--json"}, out, out); code != 0 {
		t.Fatal(out.String())
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil || payload["ok"] != true {
		t.Fatalf("list json: %s", out.String())
	}
}

// STK-REQ-260820-Y3Q4:AC-001:acceptance
func TestAcceptanceServeBoardsFromFiles(t *testing.T) {
	root := t.TempDir()
	if code := run(t, root, "init"); code != 0 {
		t.Fatal("init")
	}
	if code := run(t, root, "create", "grant", "--id", "grant_ui", "--title", "UI Grant", "--set", "status=researching"); code != 0 {
		t.Fatal("create")
	}
	a := app.OpenOrCWD(root)
	ts := httptest.NewServer(serve.Handler(a))
	t.Cleanup(ts.Close)
	res, err := http.Get(ts.URL + "/api/boards/grants")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(payload)
	if payload["ok"] != true || !strings.Contains(string(raw), "grant_ui") {
		t.Fatalf("board api: %s", raw)
	}
	home, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer home.Body.Close()
	if home.Header.Get("Content-Type") == "" {
		t.Fatal("missing ui content type")
	}
}

// STK-REQ-260820-T8AZ:AC-001:acceptance
func TestAcceptanceOneBinaryServesUI(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "crm")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/crm")
	cmd.Dir = filepath.Join("..", "..")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	if _, err := os.Stat(bin); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	init := exec.Command(bin, "--root", root, "init")
	if o, err := init.CombinedOutput(); err != nil {
		t.Fatalf("init: %v\n%s", err, o)
	}
	a := app.OpenOrCWD(root)
	ts := httptest.NewServer(serve.Handler(a))
	t.Cleanup(ts.Close)
	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("ui status %d", res.StatusCode)
	}
}
