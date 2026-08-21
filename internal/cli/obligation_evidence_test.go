package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/buger/recs/internal/cli"
)

func decodeOut(t *testing.T, raw string) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("json: %v\n%s", err, raw)
	}
	return payload
}

// Verifies: SW-REQ-260821-9737 SYS-REQ-260821-8FKR
// SW-REQ-260821-9737:empty_input:nominal
// SW-REQ-260821-9737:malformed_input:nominal
// SYS-REQ-260821-8FKR:empty_input:nominal
func TestValidSetPatchSearchSucceed(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init", errb.String())
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "create", "note", "--id", "n1", "--title", "N", "--set", "status=open", "--json"}, out, errb) != 0 {
		t.Fatal("create --set", out.String(), errb.String())
	}
	if decodeOut(t, out.String())["ok"] != true {
		t.Fatal(out.String())
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "patch", "n1", "--set", "status=done", "--json"}, out, errb) != 0 {
		t.Fatal("patch --set", out.String(), errb.String())
	}
	if decodeOut(t, out.String())["ok"] != true {
		t.Fatal(out.String())
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "search", "N", "--json"}, out, errb) != 0 {
		t.Fatal("search", out.String(), errb.String())
	}
	if decodeOut(t, out.String())["ok"] != true {
		t.Fatal(out.String())
	}
}

// Verifies: SW-REQ-260821-AY8F SYS-REQ-260821-8FKR
// SW-REQ-260821-AY8F:malformed_input:nominal
func TestKnownFlagsSucceed(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init", "--json"}, out, errb) != 0 {
		t.Fatal("init", out.String(), errb.String())
	}
	if decodeOut(t, out.String())["ok"] != true {
		t.Fatal(out.String())
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "list", "--json"}, out, errb) != 0 {
		t.Fatal("list", out.String(), errb.String())
	}
	if decodeOut(t, out.String())["ok"] != true {
		t.Fatal(out.String())
	}
}

// Verifies: SW-REQ-260821-E5V8 SYS-REQ-260821-8FKR
// SW-REQ-260821-E5V8:error_handling:nominal
func TestFoundWorkspaceNeedsNoRootHint(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "list", "--json"}, out, errb) != 0 {
		t.Fatal("list", out.String(), errb.String())
	}
	payload := decodeOut(t, out.String())
	if payload["ok"] != true {
		t.Fatal(out.String())
	}
	if msg, _ := payload["message"].(string); strings.Contains(msg, "crm.yaml not found") {
		t.Fatalf("found workspace reported missing: %#v", payload)
	}
}

// Verifies: SW-REQ-260821-MFR2 SYS-REQ-260821-8FKR
// SW-REQ-260821-MFR2:empty_input:nominal
func TestIngestFileSucceeds(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	src := filepath.Join(root, "mail.json")
	if err := os.WriteFile(src, []byte(`{"type":"email","title":"Hello","body":"hi"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "ingest", src, "--json"}, out, errb) != 0 {
		t.Fatal("ingest file", out.String(), errb.String())
	}
	if decodeOut(t, out.String())["ok"] != true {
		t.Fatal(out.String())
	}
}

// Verifies: SYS-REQ-260821-8FKR SW-REQ-260821-T9AY
// SYS-REQ-260821-8FKR:boundary:nominal
// SYS-REQ-260821-8FKR:empty_input:negative
func TestEmptyNextTriageAndEmptySearch(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "next"}, out, errb) != 0 {
		t.Fatal(errb.String())
	}
	if !strings.Contains(out.String(), "no next actions") {
		t.Fatalf("empty next: %q", out.String())
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "search", "--json"}, out, errb) == 0 {
		t.Fatal("empty search succeeded")
	}
	payload := decodeOut(t, out.String())
	if payload["ok"] != false || payload["next"] == nil {
		t.Fatalf("empty search: %#v", payload)
	}
}
