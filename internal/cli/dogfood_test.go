package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/buger/recs/internal/cli"
	"github.com/buger/recs/internal/record"
)

func decodeJSON(t *testing.T, raw string) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("json: %v\n%s", err, raw)
	}
	return payload
}

func requireErrorEnvelope(t *testing.T, payload map[string]any, code string) {
	t.Helper()
	if payload["ok"] != false {
		t.Fatalf("ok: %v", payload)
	}
	if payload["error"] != code {
		t.Fatalf("error %v want %s in %#v", payload["error"], code, payload)
	}
	if payload["message"] == nil || payload["message"] == "" {
		t.Fatalf("message missing: %#v", payload)
	}
	if payload["next"] == nil || payload["next"] == "" {
		t.Fatalf("next missing: %#v", payload)
	}
}

// Verifies: SW-REQ-260821-AY8F SYS-REQ-260821-8FKR SW-REQ-260821-FCGM INT-REQ-260821-BSH3
// SW-REQ-260821-AY8F:malformed_input:negative
// SYS-REQ-260821-8FKR:malformed_input:negative
func TestUnknownFlagsFail(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init", errb.String())
	}
	for _, args := range [][]string{
		{"--root", root, "triage", "--not-a-flag", "--json"},
		{"--root", root, "list", "--foobar", "--json"},
	} {
		out.Reset()
		errb.Reset()
		if cli.Main(args, out, errb) == 0 {
			t.Fatalf("expected fail: %v\n%s", args, out.String())
		}
		payload := decodeJSON(t, out.String())
		requireErrorEnvelope(t, payload, "unknown_flag")
		if payload["field"] != "flag" || payload["allowed"] == nil {
			t.Fatalf("allowed missing: %#v", payload)
		}
	}
}

// Verifies: SW-REQ-260821-FCGM INT-REQ-260821-BSH3 SYS-REQ-260821-8FKR
// SW-REQ-260821-FCGM:error_handling:negative
func TestEnumerableErrorsIncludeAllowed(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "board", "nope", "--json"}, out, errb) == 0 {
		t.Fatal("unknown board succeeded")
	}
	payload := decodeJSON(t, out.String())
	requireErrorEnvelope(t, payload, "unknown_board")
	if payload["field"] != "board" || payload["allowed"] == nil {
		t.Fatalf("board allowed: %#v", payload)
	}
	if !strings.Contains(payload["next"].(string), "recs board") {
		t.Fatalf("next: %#v", payload)
	}

	if cli.Main([]string{"--root", root, "create", "grant", "--id", "g1", "--title", "G"}, out, errb) != 0 {
		t.Fatal("create")
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "move", "g1", "grants", "nope-col", "--json"}, out, errb) == 0 {
		t.Fatal("unknown column succeeded")
	}
	payload = decodeJSON(t, out.String())
	if payload["error"] != "unknown_column" && payload["ok"] != false {
		t.Fatalf("column: %#v", payload)
	}
	if payload["allowed"] == nil && !strings.Contains(out.String()+errb.String(), "unknown column") {
		t.Fatalf("column recovery: %s %s", out.String(), errb.String())
	}
}

// Verifies: SW-REQ-260821-9657 SW-REQ-260820-YB5C SYS-REQ-260820-PG9C
// SW-REQ-260821-9657:boundary:nominal
// SW-REQ-260820-YB5C:boundary:nominal
func TestEmptyJSONCollectionsAreArrays(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	for _, args := range [][]string{
		{"--root", root, "next", "--json"},
		{"--root", root, "triage", "--json"},
	} {
		out.Reset()
		if cli.Main(args, out, errb) != 0 {
			t.Fatal(args, out.String(), errb.String())
		}
		raw := out.String()
		if strings.Contains(raw, `"actions": null`) || strings.Contains(raw, `"items": null`) {
			t.Fatalf("null collection: %s", raw)
		}
		payload := decodeJSON(t, raw)
		if payload["ok"] != true {
			t.Fatal(raw)
		}
	}
}

// Verifies: SW-REQ-260821-CR08 SW-REQ-260820-YB5C SYS-REQ-260820-PG9C
// SW-REQ-260821-CR08:determinism:nominal
func TestJSONTimestampsAreRFC3339(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	body := "---\nid: n1\ntype: note\ntitle: T\nupdated_at: " + time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC).Format(time.RFC3339) + "\n---\nX\n"
	if err := os.MkdirAll(filepath.Join(root, "records", "notes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "records", "notes", "n1.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "show", "n1", "--json"}, out, errb) != 0 {
		t.Fatal(out.String(), errb.String())
	}
	if strings.Contains(out.String(), " +0000 UTC") {
		t.Fatalf("Go Time.String in JSON: %s", out.String())
	}
	payload := decodeJSON(t, out.String())
	rec, _ := payload["record"].(map[string]any)
	if rec == nil {
		t.Fatal(out.String())
	}
	if ts, ok := rec["updated_at"].(string); ok {
		if _, err := time.Parse(time.RFC3339, ts); err != nil {
			t.Fatalf("not rfc3339 %q: %v", ts, err)
		}
	}
}

// Verifies: SW-REQ-260821-MFR2 SYS-REQ-260821-8FKR
// SW-REQ-260821-MFR2:empty_input:negative
func TestIngestNoFileIsUsage(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	cli.Input = strings.NewReader(`{"type":"note","title":"nope"}`)
	out.Reset()
	errb.Reset()
	if cli.Main([]string{"--root", root, "ingest", "--json"}, out, errb) == 0 {
		t.Fatal("ingest with no file should fail")
	}
	payload := decodeJSON(t, out.String())
	requireErrorEnvelope(t, payload, "error")
	if !strings.Contains(payload["message"].(string), "usage:") {
		t.Fatalf("message: %#v", payload)
	}
}

// Verifies: SW-REQ-260821-9737 SYS-REQ-260821-8FKR
// SW-REQ-260821-9737:malformed_input:negative
// SW-REQ-260821-9737:empty_input:negative
func TestSetPatchSearchUsageErrors(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	if cli.Main([]string{"--root", root, "create", "note", "--id", "n1", "--title", "N"}, out, errb) != 0 {
		t.Fatal("create")
	}
	cases := [][]string{
		{"--root", root, "create", "note", "--set", "not-an-assignment", "--json"},
		{"--root", root, "create", "note", "--json", "--set"},
		{"--root", root, "patch", "n1", "--json"},
		{"--root", root, "search", "--json"},
	}
	for _, args := range cases {
		out.Reset()
		errb.Reset()
		if cli.Main(args, out, errb) == 0 {
			t.Fatalf("expected fail: %v\n%s", args, out.String())
		}
		payload := decodeJSON(t, out.String())
		if payload["ok"] != false || payload["next"] == nil {
			t.Fatalf("%v => %#v", args, payload)
		}
	}
}

// Verifies: SW-REQ-260820-6EVX SYS-REQ-260821-8FKR SW-REQ-260821-FCGM
// SW-REQ-260820-6EVX:malformed_input:negative
func TestQueryUnknownOperatorFails(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "query", "type==grant", "--json"}, out, errb) == 0 {
		t.Fatal("== should fail", out.String())
	}
	payload := decodeJSON(t, out.String())
	requireErrorEnvelope(t, payload, "unknown_operator")
	allowed, _ := payload["allowed"].([]any)
	if len(allowed) == 0 {
		t.Fatalf("allowed: %#v", payload)
	}
	out.Reset()
	if cli.Main([]string{"help", "query"}, out, errb) != 0 {
		t.Fatal("help query")
	}
	help := out.String()
	for _, op := range []string{"=", "!=", "<", ">", "<=", ">=", "contains", "in"} {
		if !strings.Contains(help, op) {
			t.Fatalf("help query missing %s: %s", op, help)
		}
	}
}

// Verifies: SW-REQ-260821-E5V8 SYS-REQ-260821-8FKR
// SW-REQ-260821-E5V8:error_handling:negative
func TestWorkspaceAndNotFoundNext(t *testing.T) {
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	missing := filepath.Join(t.TempDir(), "no-ws")
	if err := os.MkdirAll(missing, 0o755); err != nil {
		t.Fatal(err)
	}
	if cli.Main([]string{"--root", missing, "list", "--json"}, out, errb) == 0 {
		t.Fatal("expected workspace miss")
	}
	payload := decodeJSON(t, out.String())
	requireErrorEnvelope(t, payload, "error")
	if !strings.Contains(payload["message"].(string), "--root") {
		t.Fatalf("message should mention --root: %#v", payload)
	}

	root := t.TempDir()
	out.Reset()
	errb.Reset()
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	out.Reset()
	if cli.Main([]string{"--root", root, "show", "missing", "--json"}, out, errb) == 0 {
		t.Fatal("show missing")
	}
	payload = decodeJSON(t, out.String())
	requireErrorEnvelope(t, payload, "not_found")
	if !strings.Contains(payload["next"].(string), "recs list") {
		t.Fatalf("next: %#v", payload)
	}
}

// Verifies: SW-REQ-260821-T9AY SW-REQ-260821-FCGM
// SW-REQ-260821-T9AY:boundary:nominal
func TestHumanEmptyNextTriage(t *testing.T) {
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
	if cli.Main([]string{"--root", root, "triage"}, out, errb) != 0 {
		t.Fatal(errb.String())
	}
	if !strings.Contains(out.String(), "no triage items") {
		t.Fatalf("empty triage: %q", out.String())
	}
}

// Verifies: SW-REQ-260821-FCGM SYS-REQ-260821-8FKR
// SW-REQ-260821-FCGM:nominal:nominal
func TestHelpDocumentsRecoveryFacts(t *testing.T) {
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"help", "create"}, out, errb) != 0 {
		t.Fatal("help create")
	}
	create := out.String()
	if !strings.Contains(create, "open-ended") || !strings.Contains(create, "--root") {
		t.Fatalf("create help: %s", create)
	}
	out.Reset()
	if cli.Main([]string{"help", "next"}, out, errb) != 0 {
		t.Fatal("help next")
	}
	if !strings.Contains(out.String(), "next_action") || !strings.Contains(out.String(), "priority") {
		t.Fatalf("next help: %s", out.String())
	}
	out.Reset()
	if cli.Main([]string{"help", "triage"}, out, errb) != 0 {
		t.Fatal("help triage")
	}
	triage := out.String()
	for _, reason := range []string{"inbox", "overdue", "blocker", "missing_metadata"} {
		if !strings.Contains(triage, reason) {
			t.Fatalf("triage help missing %s: %s", reason, triage)
		}
	}
	out.Reset()
	if cli.Main([]string{"help", "ingest"}, out, errb) != 0 {
		t.Fatal("help ingest")
	}
	ingest := out.String()
	if !strings.Contains(ingest, `"type":"email"`) && !strings.Contains(ingest, `"type": "email"`) {
		t.Fatalf("ingest example: %s", ingest)
	}
	out.Reset()
	if cli.Main([]string{"help", "move"}, out, errb) != 0 {
		t.Fatal("help move")
	}
	move := out.String()
	if strings.Contains(move, "grant_1") {
		t.Fatalf("move help should use placeholder: %s", move)
	}
	if !strings.Contains(move, "<id>") {
		t.Fatalf("move help placeholder: %s", move)
	}
	out.Reset()
	if cli.Main([]string{"help", "export"}, out, errb) != 0 {
		t.Fatal("help export")
	}
	if !strings.Contains(out.String(), "flat subset") {
		t.Fatalf("export help: %s", out.String())
	}
	out.Reset()
	if cli.Main([]string{"help", "list"}, out, errb) != 0 {
		t.Fatal("help list")
	}
	if !strings.Contains(out.String(), "--root") && !strings.Contains(out.String(), "Global flags") {
		t.Fatalf("list help should mention --root or global flags: %s", out.String())
	}
}

// Verifies: SW-REQ-260821-CR08
func TestGetStringTimeIsRFC3339(t *testing.T) {
	rec := &record.Record{Fields: map[string]any{"due": time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}}
	got := rec.GetString("due")
	if _, err := time.Parse(time.RFC3339, got); err != nil {
		t.Fatalf("GetString time: %q", got)
	}
	if record.TypeDir("inbox") != "inbox" {
		t.Fatalf("inbox folder: %s", record.TypeDir("inbox"))
	}
}
