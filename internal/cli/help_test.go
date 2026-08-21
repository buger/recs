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

// Verifies: SYS-REQ-260821-8FKR SW-REQ-260821-FCGM INT-REQ-260821-BSH3 STK-REQ-260820-V5ZD
// SYS-REQ-260821-8FKR:nominal:nominal
// SYS-REQ-260821-8FKR:error_handling:nominal
// SYS-REQ-260821-8FKR:error_handling:negative
// SW-REQ-260821-FCGM:nominal:nominal
// SW-REQ-260821-FCGM:error_handling:nominal
// SW-REQ-260821-FCGM:error_handling:negative
// INT-REQ-260821-BSH3:nominal:nominal
// INT-REQ-260821-BSH3:integration:integration
// MCDC SW-REQ-260821-FCGM: agent_discovery_requested=T, agent_sidecar_written=F, arg_count_GE_1=T, command_help_emitted=F, command_rejected=F, global_help_emitted=T, structured_error_emitted=F => TRUE
// MCDC SW-REQ-260821-FCGM: agent_discovery_requested=T, agent_sidecar_written=F, arg_count_GE_1=T, command_help_emitted=F, command_rejected=F, global_help_emitted=F, structured_error_emitted=T => TRUE
// MCDC INT-REQ-260821-BSH3: agent_discovery_requested=T, arg_count_GT_0=T, help_text_emitted=T, sidecar_file_required=F, structured_error_emitted=F => TRUE
// MCDC INT-REQ-260821-BSH3: agent_discovery_requested=T, arg_count_GT_0=T, help_text_emitted=F, sidecar_file_required=F, structured_error_emitted=T => TRUE
//mcdc:ignore SYS-REQ-260821-8FKR: agent_discovery_requested=T, arg_count_GE_0=T, command_help_emitted=F, command_rejected=F, global_help_emitted=F, structured_error_emitted=F => FALSE -- discovery without help or structured error is the literal negation of the agent-interface contract [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260821-FCGM: agent_discovery_requested=T, agent_sidecar_written=F, arg_count_GE_1=T, command_help_emitted=F, command_rejected=F, global_help_emitted=F, structured_error_emitted=F => FALSE -- discovery without help or structured error is the literal negation of the agent-interface contract [reviewed: agent:grok] [category: defensive]
//mcdc:ignore INT-REQ-260821-BSH3: agent_discovery_requested=T, arg_count_GT_0=T, help_text_emitted=F, sidecar_file_required=F, structured_error_emitted=F => FALSE -- discovery without help or structured error is the literal negation of the agent-interface contract [reviewed: agent:grok] [category: defensive]
func TestHelpIsTheAgentInterface(t *testing.T) {
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main(nil, out, errb) != 0 {
		t.Fatal("empty")
	}
	text := out.String()
	if !strings.Contains(text, "recs help") || strings.Contains(text, "Write AGENTS.md") {
		t.Fatalf("global help: %s", text)
	}
	for _, name := range []string{"init", "create", "show", "list", "search", "query", "set", "patch", "board", "dashboard", "move", "next", "triage", "validate", "index", "context", "inbox", "serve", "edit", "delete", "link", "ingest", "export", "import", "diff", "changed", "history", "help"} {
		if !strings.Contains(text, name) {
			t.Fatalf("missing command %s in --help", name)
		}
	}

	out.Reset()
	if cli.Main([]string{"--help"}, out, errb) != 0 || !strings.Contains(out.String(), "query") {
		t.Fatal("recs --help")
	}
	out.Reset()
	if cli.Main([]string{"-h"}, out, errb) != 0 {
		t.Fatal("-h")
	}
	out.Reset()
	if cli.Main([]string{"help"}, out, errb) != 0 || !strings.Contains(out.String(), "Commands:") {
		t.Fatal("recs help")
	}

	out.Reset()
	if cli.Main([]string{"help", "query"}, out, errb) != 0 {
		t.Fatal("help query")
	}
	q := out.String()
	if !strings.Contains(q, "recs query") || !strings.Contains(q, "JSON shape") || !strings.Contains(q, "--json") {
		t.Fatalf("command help: %s", q)
	}

	out.Reset()
	if cli.Main([]string{"query", "--help"}, out, errb) != 0 || !strings.Contains(out.String(), "Examples:") {
		t.Fatal("query --help")
	}
	out.Reset()
	if cli.Main([]string{"create", "-h"}, out, errb) != 0 || !strings.Contains(out.String(), "--title") {
		t.Fatal("create -h")
	}

	out.Reset()
	if cli.Main([]string{"--json", "--help"}, out, errb) != 0 {
		t.Fatal("json help")
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil || payload["ok"] != true {
		t.Fatalf("json help: %s", out.String())
	}
	cmds, _ := payload["commands"].([]any)
	if len(cmds) < 20 {
		t.Fatalf("json command list short: %d", len(cmds))
	}

	out.Reset()
	if cli.Main([]string{"help", "set", "--json"}, out, errb) != 0 {
		t.Fatal("help set json")
	}
	payload = map[string]any{}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil || payload["command"] != "set" {
		t.Fatalf("help set json: %s", out.String())
	}

	out.Reset()
	errb.Reset()
	if cli.Main([]string{"help", "nope", "--json"}, out, errb) != 1 {
		t.Fatal("unknown help topic")
	}
	payload = map[string]any{}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ok"] != false || payload["error"] != "unknown_command" || payload["next"] != "recs --help" {
		t.Fatalf("unknown help json: %s", out.String())
	}
	if payload["field"] != "command" || payload["allowed"] == nil {
		t.Fatalf("allowed missing: %s", out.String())
	}

	out.Reset()
	errb.Reset()
	if cli.Main([]string{"help", "nope"}, out, errb) != 1 || !strings.Contains(errb.String(), "next: recs --help") {
		t.Fatalf("unknown help human: %s", errb.String())
	}

	out.Reset()
	errb.Reset()
	if cli.Main([]string{"agent", "install", "--json"}, out, errb) != 1 {
		t.Fatal("agent install must fail")
	}
	payload = map[string]any{}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["error"] != "unknown_command" || payload["next"] != "recs --help" {
		t.Fatalf("agent json: %s", out.String())
	}
	allowed, _ := payload["allowed"].([]any)
	for _, a := range allowed {
		if a == "agent" {
			t.Fatal("agent still listed")
		}
	}

	root := t.TempDir()
	out.Reset()
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	out.Reset()
	errb.Reset()
	if cli.Main([]string{"--root", root, "show", "--json"}, out, errb) != 1 {
		t.Fatal("usage")
	}
	payload = map[string]any{}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["next"] != "recs help show" {
		t.Fatalf("usage next: %s", out.String())
	}
	for _, name := range []string{"AGENTS.md", "SKILL.md"} {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			t.Fatalf("init wrote %s", name)
		}
	}

	out.Reset()
	errb.Reset()
	if cli.Main([]string{"--root", root, "set", "x", "title", "Z"}, out, errb) == 0 {
		t.Fatal("expected fail")
	}
	if !strings.Contains(errb.String(), "next:") && !strings.Contains(out.String(), "next") {
		t.Fatalf("human next missing: %s %s", out.String(), errb.String())
	}
}

// Verifies: SYS-REQ-260821-8FKR SW-REQ-260821-FCGM
func TestHelpCoversEveryCatalogCommand(t *testing.T) {
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--json", "help"}, out, errb) != 0 {
		t.Fatal(out.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	cmds, _ := payload["commands"].([]any)
	for _, raw := range cmds {
		item, _ := raw.(map[string]any)
		name, _ := item["name"].(string)
		if name == "" {
			t.Fatal(raw)
		}
		out.Reset()
		if cli.Main([]string{"help", name, "--json"}, out, errb) != 0 {
			t.Fatalf("help %s: %s", name, out.String())
		}
		var one map[string]any
		if err := json.Unmarshal(out.Bytes(), &one); err != nil || one["command"] != name {
			t.Fatalf("help %s json: %s", name, out.String())
		}
	}
}
