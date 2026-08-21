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

	"github.com/buger/recs/internal/app"
	"github.com/buger/recs/internal/cli"
	"github.com/buger/recs/internal/serve"
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
	bin := filepath.Join(t.TempDir(), "recs")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/recs")
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

// STK-REQ-260820-V5ZD:AC-002:acceptance
func TestAcceptanceHelpAndJSONError(t *testing.T) {
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--help"}, out, errb) != 0 {
		t.Fatal(out.String(), errb.String())
	}
	text := out.String()
	for _, name := range []string{"init", "create", "help", "query", "serve"} {
		if !strings.Contains(text, name) {
			t.Fatalf("help missing %s: %s", name, text)
		}
	}
	out.Reset()
	errb.Reset()
	code := cli.Main([]string{"--json", "not-a-command"}, out, errb)
	if code == 0 {
		t.Fatal("unknown command succeeded")
	}
	raw := errb.Bytes()
	if len(raw) == 0 {
		raw = out.Bytes()
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("json: %v %s", err, raw)
	}
	if payload["ok"] != false || payload["error"] == nil || payload["message"] == nil {
		t.Fatalf("shape: %#v", payload)
	}
}

// STK-REQ-260821-QTPP:AC-001:acceptance
func TestAcceptanceRemainingMutationsJSON(t *testing.T) {
	root := t.TempDir()
	if run(t, root, "init") != 0 {
		t.Fatal("init")
	}
	if run(t, root, "create", "person", "--id", "person_alice", "--name", "Alice") != 0 {
		t.Fatal("create")
	}
	if run(t, root, "create", "company", "--id", "company_acme", "--name", "Acme") != 0 {
		t.Fatal("company")
	}
	if run(t, root, "link", "person_alice", "company_acme", "--relation", "works_at", "--json") != 0 {
		t.Fatal("link")
	}
	out := &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "export", "--json"}, out, out) != 0 {
		t.Fatal(out.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil || payload["ok"] != true {
		t.Fatalf("export: %s", out.String())
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "delete", "missing", "--json"}, out, out) == 0 {
		t.Fatal("missing delete succeeded")
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil || payload["ok"] == true {
		t.Fatalf("delete error: %s", out.String())
	}
}

// STK-REQ-260821-NTWY:AC-001:acceptance
func TestAcceptanceRecordViewEditorSearch(t *testing.T) {
	root := t.TempDir()
	if run(t, root, "init") != 0 {
		t.Fatal("init")
	}
	if run(t, root, "create", "person", "--id", "person_alice", "--name", "Alice Smith") != 0 {
		t.Fatal("create")
	}
	a := app.OpenOrCWD(root)
	ts := httptest.NewServer(serve.Handler(a))
	t.Cleanup(ts.Close)
	for _, path := range []string{"/", "/api/records/person_alice", "/api/search?q=Alice"} {
		res, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatal(path, err)
		}
		res.Body.Close()
		if res.StatusCode != 200 {
			t.Fatalf("%s status %d", path, res.StatusCode)
		}
	}
}

// STK-REQ-260820-4255:AC-001:acceptance
func TestAcceptanceDashboardGallery(t *testing.T) {
	root := t.TempDir()
	if run(t, root, "init") != 0 {
		t.Fatal("init")
	}
	a := app.OpenOrCWD(root)
	ts := httptest.NewServer(serve.Handler(a))
	t.Cleanup(ts.Close)
	home, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer home.Body.Close()
	buf := &bytes.Buffer{}
	buf.ReadFrom(home.Body)
	body := buf.String()
	if home.StatusCode != 200 || (!strings.Contains(body, "dashboard") && !strings.Contains(body, "prospects") && !strings.Contains(strings.ToLower(body), "workspace")) {
		api, err := http.Get(ts.URL + "/api/dashboards")
		if err != nil {
			t.Fatal(err)
		}
		defer api.Body.Close()
		var payload map[string]any
		if err := json.NewDecoder(api.Body).Decode(&payload); err != nil {
			t.Fatalf("gallery ui=%q api=%v", body[:min(200, len(body))], err)
		}
		if payload["ok"] != true {
			t.Fatalf("dashboards api: %#v ui=%s", payload, body[:min(300, len(body))])
		}
	}
}

// STK-REQ-260820-V5ZD:AC-003:acceptance
func TestAcceptanceUnknownFlagAndIncompleteCommand(t *testing.T) {
	root := t.TempDir()
	if run(t, root, "init") != 0 {
		t.Fatal("init")
	}

	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "list", "--not-a-flag", "--json"}, out, errb) == 0 {
		t.Fatal("unknown flag succeeded")
	}
	raw := out.Bytes()
	if len(raw) == 0 {
		raw = errb.Bytes()
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unknown flag json: %v %s", err, raw)
	}
	if payload["ok"] != false {
		t.Fatalf("unknown flag ok: %#v", payload)
	}
	next, _ := payload["next"].(string)
	if next == "" {
		t.Fatalf("unknown flag missing next: %#v", payload)
	}
	if !strings.Contains(next, "recs help") && !strings.Contains(next, "recs --help") {
		t.Fatalf("unknown flag next is not a recovery command: %#v", payload)
	}

	out.Reset()
	errb.Reset()
	if cli.Main([]string{"--root", root, "show", "--json"}, out, errb) == 0 {
		t.Fatal("incomplete command succeeded")
	}
	raw = out.Bytes()
	if len(raw) == 0 {
		raw = errb.Bytes()
	}
	payload = map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("incomplete json: %v %s", err, raw)
	}
	if payload["ok"] != false {
		t.Fatalf("incomplete ok: %#v", payload)
	}
	next, _ = payload["next"].(string)
	if next == "" {
		t.Fatalf("incomplete missing next: %#v", payload)
	}
	if !strings.Contains(next, "recs help") && !strings.Contains(next, "recs --help") {
		t.Fatalf("incomplete next is not a recovery command: %#v", payload)
	}
}
