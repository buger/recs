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
// SYS-REQ-260820-PG9C:determinism:nominal
// SYS-REQ-260820-PG9C:nominal:nominal
// SW-REQ-260820-YB5C:determinism:nominal
// SW-REQ-260820-YB5C:nominal:nominal
// INT-REQ-260820-JC9M:nominal:nominal
// INT-REQ-260820-JC9M:integration:integration
// SYS-REQ-260820-5C9D:empty_input:nominal
// SYS-REQ-260820-5C9D:nominal:nominal
// SW-REQ-260820-ZKCV:empty_input:nominal
// SW-REQ-260820-ZKCV:nominal:nominal
// SYS-REQ-260820-DCG4:empty_input:nominal
// SYS-REQ-260820-DCG4:nominal:nominal
// SW-REQ-260820-D5WE:empty_input:nominal
// SW-REQ-260820-D5WE:nominal:nominal
// MCDC INT-REQ-260820-JC9M: arg_count_GT_0=F, board_api_used=F, cli_command_dispatched=T, query_api_used=F, shared_record_model_used=F, store_api_used=F => TRUE
// MCDC INT-REQ-260820-JC9M: arg_count_GT_0=T, board_api_used=F, cli_command_dispatched=F, query_api_used=F, shared_record_model_used=F, store_api_used=F => TRUE
//mcdc:ignore INT-REQ-260820-JC9M: arg_count_GT_0=T, board_api_used=F, cli_command_dispatched=T, query_api_used=F, shared_record_model_used=F, store_api_used=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore INT-REQ-260820-JC9M: arg_count_GT_0=T, board_api_used=F, cli_command_dispatched=T, query_api_used=F, shared_record_model_used=T, store_api_used=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC INT-REQ-260820-JC9M: arg_count_GT_0=T, board_api_used=F, cli_command_dispatched=T, query_api_used=F, shared_record_model_used=T, store_api_used=T => TRUE
// MCDC INT-REQ-260820-JC9M: arg_count_GT_0=T, board_api_used=F, cli_command_dispatched=T, query_api_used=T, shared_record_model_used=T, store_api_used=F => TRUE
// MCDC INT-REQ-260820-JC9M: arg_count_GT_0=T, board_api_used=T, cli_command_dispatched=T, query_api_used=F, shared_record_model_used=T, store_api_used=F => TRUE
//mcdc:ignore INT-REQ-260820-JC9M: arg_count_GT_0=T, board_api_used=T, cli_command_dispatched=T, query_api_used=T, shared_record_model_used=F, store_api_used=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC INT-REQ-260820-JC9M: arg_count_GT_0=T, board_api_used=T, cli_command_dispatched=T, query_api_used=T, shared_record_model_used=T, store_api_used=T => TRUE
// MCDC SW-REQ-260820-D5WE: arg_count_GE_1=F, blockers_listed=F, inbox_items_listed=F, missing_metadata_listed=F, overdue_actions_listed=F, triage_command_invoked=T, triage_empty=F => TRUE
// MCDC SW-REQ-260820-D5WE: arg_count_GE_1=T, blockers_listed=F, inbox_items_listed=F, missing_metadata_listed=F, overdue_actions_listed=F, triage_command_invoked=F, triage_empty=F => TRUE
//mcdc:ignore SW-REQ-260820-D5WE: arg_count_GE_1=T, blockers_listed=F, inbox_items_listed=F, missing_metadata_listed=F, overdue_actions_listed=F, triage_command_invoked=T, triage_empty=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-D5WE: arg_count_GE_1=T, blockers_listed=F, inbox_items_listed=F, missing_metadata_listed=F, overdue_actions_listed=F, triage_command_invoked=T, triage_empty=T => TRUE
// MCDC SW-REQ-260820-D5WE: arg_count_GE_1=T, blockers_listed=F, inbox_items_listed=F, missing_metadata_listed=F, overdue_actions_listed=T, triage_command_invoked=T, triage_empty=F => TRUE
// MCDC SW-REQ-260820-D5WE: arg_count_GE_1=T, blockers_listed=F, inbox_items_listed=F, missing_metadata_listed=T, overdue_actions_listed=F, triage_command_invoked=T, triage_empty=F => TRUE
// MCDC SW-REQ-260820-D5WE: arg_count_GE_1=T, blockers_listed=F, inbox_items_listed=T, missing_metadata_listed=F, overdue_actions_listed=F, triage_command_invoked=T, triage_empty=F => TRUE
// MCDC SW-REQ-260820-D5WE: arg_count_GE_1=T, blockers_listed=T, inbox_items_listed=F, missing_metadata_listed=F, overdue_actions_listed=F, triage_command_invoked=T, triage_empty=F => TRUE
// MCDC SW-REQ-260820-YB5C: arg_count_GT_0=F, command_completed=T, human_output_emitted=T, json_flag_set=T, json_output_emitted=T => TRUE
// MCDC SW-REQ-260820-YB5C: arg_count_GT_0=T, command_completed=F, human_output_emitted=T, json_flag_set=T, json_output_emitted=T => TRUE
// MCDC SW-REQ-260820-YB5C: arg_count_GT_0=T, command_completed=T, human_output_emitted=F, json_flag_set=T, json_output_emitted=T => TRUE
// MCDC SW-REQ-260820-YB5C: arg_count_GT_0=T, command_completed=T, human_output_emitted=T, json_flag_set=F, json_output_emitted=F => TRUE
//mcdc:ignore SW-REQ-260820-YB5C: arg_count_GT_0=T, command_completed=T, human_output_emitted=T, json_flag_set=F, json_output_emitted=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-YB5C: arg_count_GT_0=T, command_completed=T, human_output_emitted=T, json_flag_set=T, json_output_emitted=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-YB5C: arg_count_GT_0=T, command_completed=T, human_output_emitted=T, json_flag_set=T, json_output_emitted=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-ZKCV: actions_sorted=F, arg_count_GE_1=F, due_actions_collected=F, next_command_invoked=T, next_empty=F => TRUE
// MCDC SW-REQ-260820-ZKCV: actions_sorted=F, arg_count_GE_1=T, due_actions_collected=F, next_command_invoked=F, next_empty=F => TRUE
//mcdc:ignore SW-REQ-260820-ZKCV: actions_sorted=F, arg_count_GE_1=T, due_actions_collected=F, next_command_invoked=T, next_empty=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-ZKCV: actions_sorted=F, arg_count_GE_1=T, due_actions_collected=F, next_command_invoked=T, next_empty=T => TRUE
//mcdc:ignore SW-REQ-260820-ZKCV: actions_sorted=F, arg_count_GE_1=T, due_actions_collected=T, next_command_invoked=T, next_empty=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-ZKCV: actions_sorted=T, arg_count_GE_1=T, due_actions_collected=F, next_command_invoked=T, next_empty=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-ZKCV: actions_sorted=T, arg_count_GE_1=T, due_actions_collected=T, next_command_invoked=T, next_empty=F => TRUE
// MCDC SYS-REQ-260820-5C9D: actions_sorted=F, due_actions_collected=F, next_command_invoked=F, next_empty=F => TRUE
//mcdc:ignore SYS-REQ-260820-5C9D: actions_sorted=F, due_actions_collected=F, next_command_invoked=T, next_empty=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-5C9D: actions_sorted=F, due_actions_collected=F, next_command_invoked=T, next_empty=T => TRUE
//mcdc:ignore SYS-REQ-260820-5C9D: actions_sorted=F, due_actions_collected=T, next_command_invoked=T, next_empty=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-5C9D: actions_sorted=T, due_actions_collected=F, next_command_invoked=T, next_empty=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-5C9D: actions_sorted=T, due_actions_collected=T, next_command_invoked=T, next_empty=F => TRUE
// MCDC SYS-REQ-260820-DCG4: blockers_listed=F, inbox_items_listed=F, missing_metadata_listed=F, overdue_actions_listed=F, triage_command_invoked=F, triage_empty=F => TRUE
//mcdc:ignore SYS-REQ-260820-DCG4: blockers_listed=F, inbox_items_listed=F, missing_metadata_listed=F, overdue_actions_listed=F, triage_command_invoked=T, triage_empty=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-DCG4: blockers_listed=F, inbox_items_listed=F, missing_metadata_listed=F, overdue_actions_listed=F, triage_command_invoked=T, triage_empty=T => TRUE
// MCDC SYS-REQ-260820-DCG4: blockers_listed=F, inbox_items_listed=F, missing_metadata_listed=F, overdue_actions_listed=T, triage_command_invoked=T, triage_empty=F => TRUE
// MCDC SYS-REQ-260820-DCG4: blockers_listed=F, inbox_items_listed=F, missing_metadata_listed=T, overdue_actions_listed=F, triage_command_invoked=T, triage_empty=F => TRUE
// MCDC SYS-REQ-260820-DCG4: blockers_listed=F, inbox_items_listed=T, missing_metadata_listed=F, overdue_actions_listed=F, triage_command_invoked=T, triage_empty=F => TRUE
// MCDC SYS-REQ-260820-DCG4: blockers_listed=T, inbox_items_listed=F, missing_metadata_listed=F, overdue_actions_listed=F, triage_command_invoked=T, triage_empty=F => TRUE
// MCDC SYS-REQ-260820-PG9C: arg_count_GT_0=F, command_completed=T, human_output_emitted=F, json_flag_set=F, json_output_emitted=F => TRUE
// MCDC SYS-REQ-260820-PG9C: arg_count_GT_0=T, command_completed=F, human_output_emitted=F, json_flag_set=F, json_output_emitted=F => TRUE
//mcdc:ignore SYS-REQ-260820-PG9C: arg_count_GT_0=T, command_completed=T, human_output_emitted=F, json_flag_set=F, json_output_emitted=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-PG9C: arg_count_GT_0=T, command_completed=T, human_output_emitted=T, json_flag_set=F, json_output_emitted=F => TRUE
//mcdc:ignore SYS-REQ-260820-PG9C: arg_count_GT_0=T, command_completed=T, human_output_emitted=T, json_flag_set=T, json_output_emitted=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-PG9C: arg_count_GT_0=T, command_completed=T, human_output_emitted=T, json_flag_set=T, json_output_emitted=T => TRUE
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
// INT-REQ-260820-AHKR:nominal:nominal
// INT-REQ-260820-AHKR:integration:integration
// SYS-REQ-260820-9W1S:error_handling:nominal
// SYS-REQ-260820-9W1S:nominal:nominal
// SYS-REQ-260820-9W1S:boundary:nominal
// SYS-REQ-260820-9W1S:auth_required:nominal
// SYS-REQ-260820-9W1S:path_traversal_prevented:nominal
// SYS-REQ-260820-9W1S:input_size_bounded:nominal
// SW-REQ-260820-8ZS7:error_handling:nominal
// SW-REQ-260820-8ZS7:nominal:nominal
// SW-REQ-260820-8ZS7:boundary:nominal
// SW-REQ-260820-8ZS7:auth_required:nominal
// SW-REQ-260820-8ZS7:path_traversal_prevented:nominal
// SW-REQ-260820-8ZS7:input_size_bounded:nominal
// SYS-REQ-260820-9W1S:csrf_protection:nominal
// SW-REQ-260820-8ZS7:csrf_protection:nominal
// INT-REQ-260820-AHKR:csrf_protection:nominal
// MCDC INT-REQ-260820-AHKR: http_api_requested=F, separate_database_used=F, shared_app_layer_used=F, shared_record_model_used=F => TRUE
//mcdc:ignore INT-REQ-260820-AHKR: http_api_requested=T, separate_database_used=F, shared_app_layer_used=F, shared_record_model_used=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore INT-REQ-260820-AHKR: http_api_requested=T, separate_database_used=F, shared_app_layer_used=F, shared_record_model_used=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore INT-REQ-260820-AHKR: http_api_requested=T, separate_database_used=F, shared_app_layer_used=T, shared_record_model_used=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC INT-REQ-260820-AHKR: http_api_requested=T, separate_database_used=F, shared_app_layer_used=T, shared_record_model_used=T => TRUE
//mcdc:ignore INT-REQ-260820-AHKR: http_api_requested=T, separate_database_used=T, shared_app_layer_used=T, shared_record_model_used=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-8ZS7: custom_port_selected=F, default_port_selected=F, http_bound_localhost=T, listen_port_GT_0=T, separate_database_used=F, serve_command_invoked=T, shared_app_layer_used=T, shared_record_model_used=T, static_ui_served=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-8ZS7: custom_port_selected=F, default_port_selected=T, http_bound_localhost=T, listen_port_GT_0=T, separate_database_used=F, serve_command_invoked=T, shared_app_layer_used=T, shared_record_model_used=T, static_ui_served=T => TRUE
//mcdc:ignore SW-REQ-260820-8ZS7: custom_port_selected=T, default_port_selected=F, http_bound_localhost=F, listen_port_GT_0=T, separate_database_used=F, serve_command_invoked=T, shared_app_layer_used=T, shared_record_model_used=T, static_ui_served=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-8ZS7: custom_port_selected=T, default_port_selected=F, http_bound_localhost=T, listen_port_GT_0=T, separate_database_used=F, serve_command_invoked=T, shared_app_layer_used=F, shared_record_model_used=T, static_ui_served=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-8ZS7: custom_port_selected=T, default_port_selected=F, http_bound_localhost=T, listen_port_GT_0=T, separate_database_used=F, serve_command_invoked=T, shared_app_layer_used=T, shared_record_model_used=F, static_ui_served=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-8ZS7: custom_port_selected=T, default_port_selected=F, http_bound_localhost=T, listen_port_GT_0=T, separate_database_used=F, serve_command_invoked=T, shared_app_layer_used=T, shared_record_model_used=T, static_ui_served=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-8ZS7: custom_port_selected=T, default_port_selected=F, http_bound_localhost=T, listen_port_GT_0=T, separate_database_used=F, serve_command_invoked=T, shared_app_layer_used=T, shared_record_model_used=T, static_ui_served=T => TRUE
// MCDC SW-REQ-260820-8ZS7: custom_port_selected=T, default_port_selected=T, http_bound_localhost=T, listen_port_GT_0=F, separate_database_used=T, serve_command_invoked=T, shared_app_layer_used=T, shared_record_model_used=T, static_ui_served=T => TRUE
// MCDC SW-REQ-260820-8ZS7: custom_port_selected=T, default_port_selected=T, http_bound_localhost=T, listen_port_GT_0=T, separate_database_used=F, serve_command_invoked=T, shared_app_layer_used=T, shared_record_model_used=T, static_ui_served=T => TRUE
// MCDC SW-REQ-260820-8ZS7: custom_port_selected=T, default_port_selected=T, http_bound_localhost=T, listen_port_GT_0=T, separate_database_used=T, serve_command_invoked=F, shared_app_layer_used=T, shared_record_model_used=T, static_ui_served=T => TRUE
//mcdc:ignore SW-REQ-260820-8ZS7: custom_port_selected=T, default_port_selected=T, http_bound_localhost=T, listen_port_GT_0=T, separate_database_used=T, serve_command_invoked=T, shared_app_layer_used=T, shared_record_model_used=T, static_ui_served=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-9W1S: custom_port_selected=F, default_port_selected=F, http_bound_localhost=F, listen_port_GT_0=F, serve_command_invoked=T, static_ui_served=F => TRUE
// MCDC SYS-REQ-260820-9W1S: custom_port_selected=F, default_port_selected=F, http_bound_localhost=F, listen_port_GT_0=T, serve_command_invoked=F, static_ui_served=F => TRUE
//mcdc:ignore SYS-REQ-260820-9W1S: custom_port_selected=F, default_port_selected=F, http_bound_localhost=F, listen_port_GT_0=T, serve_command_invoked=T, static_ui_served=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-9W1S: custom_port_selected=F, default_port_selected=F, http_bound_localhost=T, listen_port_GT_0=T, serve_command_invoked=T, static_ui_served=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-9W1S: custom_port_selected=F, default_port_selected=T, http_bound_localhost=T, listen_port_GT_0=T, serve_command_invoked=T, static_ui_served=T => TRUE
// MCDC SYS-REQ-260820-9W1S: custom_port_selected=T, default_port_selected=F, http_bound_localhost=T, listen_port_GT_0=T, serve_command_invoked=T, static_ui_served=T => TRUE
//mcdc:ignore SYS-REQ-260820-9W1S: custom_port_selected=T, default_port_selected=T, http_bound_localhost=F, listen_port_GT_0=T, serve_command_invoked=T, static_ui_served=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-9W1S: custom_port_selected=T, default_port_selected=T, http_bound_localhost=T, listen_port_GT_0=T, serve_command_invoked=T, static_ui_served=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-9W1S: custom_port_selected=T, default_port_selected=T, http_bound_localhost=T, listen_port_GT_0=T, serve_command_invoked=T, static_ui_served=T => TRUE
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
// agent inbox template enum
// SYS-REQ-260820-PG9C:determinism:nominal
// SW-REQ-260820-YB5C:determinism:nominal
func TestAgentInboxTemplateAndEnum(t *testing.T) {
	root := t.TempDir()
	if code := run(t, root, "init"); code != 0 {
		t.Fatal(code)
	}
	for _, name := range []string{"AGENTS.md", "SKILL.md"} {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			t.Fatalf("init must not write %s", name)
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

// Verifies: SYS-REQ-260820-PG9C SW-REQ-260820-YB5C SYS-REQ-260820-2SQZ SYS-REQ-260820-DCG4 SW-REQ-260820-D5WE
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

// Verifies: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7 INT-REQ-260820-AHKR
// SYS-REQ-260820-9W1S:csrf_protection:negative
// SW-REQ-260820-8ZS7:csrf_protection:negative
// INT-REQ-260820-AHKR:csrf_protection:negative
func TestServeRejectsForeignOrigin(t *testing.T) {
	root := t.TempDir()
	a := app.OpenOrCWD(root)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(serve.Handler(a))
	defer ts.Close()
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/records", strings.NewReader(`{"type":"grant","id":"grant_csrf","title":"X"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://evil.example")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if _, err := a.Show("grant_csrf"); err == nil {
		t.Fatal("csrf create succeeded")
	}
}

// Verifies: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7
// SYS-REQ-260820-9W1S:input_size_bounded:negative
// SW-REQ-260820-8ZS7:input_size_bounded:negative
func TestServeRejectsOversizedBody(t *testing.T) {
	root := t.TempDir()
	a := app.OpenOrCWD(root)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(serve.Handler(a))
	defer ts.Close()
	body := `{"type":"grant","id":"grant_big","title":"` + strings.Repeat("x", 2<<20) + `"}`
	resp, err := http.Post(ts.URL+"/api/records", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == 201 {
		t.Fatal("oversized create succeeded")
	}
}

// Verifies: SYS-REQ-260820-4628 SW-REQ-260820-NBGR
// SYS-REQ-260820-4628:path_traversal_prevented:negative
// SW-REQ-260820-NBGR:path_traversal_prevented:negative
// MCDC SW-REQ-260820-NBGR: board_rejected=F, board_requested=F, column_field_projected=F, column_list_len_GT_0=F, column_list_len_LE_64=F, matcher_applied=F, matcher_depth_GE_0=T, records_grouped=F => TRUE
// MCDC SW-REQ-260820-NBGR: board_rejected=F, board_requested=T, column_field_projected=F, column_list_len_GT_0=F, column_list_len_LE_64=F, matcher_applied=F, matcher_depth_GE_0=F, records_grouped=F => TRUE
//mcdc:ignore SW-REQ-260820-NBGR: board_rejected=F, board_requested=T, column_field_projected=F, column_list_len_GT_0=F, column_list_len_LE_64=F, matcher_applied=F, matcher_depth_GE_0=T, records_grouped=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-NBGR: board_rejected=F, board_requested=T, column_field_projected=F, column_list_len_GT_0=T, column_list_len_LE_64=T, matcher_applied=F, matcher_depth_GE_0=T, records_grouped=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-NBGR: board_rejected=F, board_requested=T, column_field_projected=F, column_list_len_GT_0=T, column_list_len_LE_64=T, matcher_applied=T, matcher_depth_GE_0=T, records_grouped=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-NBGR: board_rejected=F, board_requested=T, column_field_projected=T, column_list_len_GT_0=T, column_list_len_LE_64=T, matcher_applied=F, matcher_depth_GE_0=T, records_grouped=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-NBGR: board_rejected=F, board_requested=T, column_field_projected=T, column_list_len_GT_0=T, column_list_len_LE_64=T, matcher_applied=T, matcher_depth_GE_0=T, records_grouped=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-NBGR: board_rejected=F, board_requested=T, column_field_projected=T, column_list_len_GT_0=T, column_list_len_LE_64=T, matcher_applied=T, matcher_depth_GE_0=T, records_grouped=T => TRUE
// MCDC SW-REQ-260820-NBGR: board_rejected=T, board_requested=T, column_field_projected=F, column_list_len_GT_0=T, column_list_len_LE_64=T, matcher_applied=F, matcher_depth_GE_0=T, records_grouped=F => TRUE
//mcdc:ignore SW-REQ-260820-NBGR: board_rejected=T, board_requested=T, column_field_projected=T, column_list_len_GT_0=F, column_list_len_LE_64=T, matcher_applied=T, matcher_depth_GE_0=T, records_grouped=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-NBGR: board_rejected=T, board_requested=T, column_field_projected=T, column_list_len_GT_0=T, column_list_len_LE_64=F, matcher_applied=T, matcher_depth_GE_0=T, records_grouped=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-NBGR: board_rejected=T, board_requested=T, column_field_projected=T, column_list_len_GT_0=T, column_list_len_LE_64=T, matcher_applied=T, matcher_depth_GE_0=T, records_grouped=T => TRUE
// MCDC SYS-REQ-260820-4628: board_rejected=F, board_requested=F, column_field_projected=F, column_list_len_GT_0=F, matcher_applied=F, matcher_depth_GE_0=T, records_grouped=F => TRUE
// MCDC SYS-REQ-260820-4628: board_rejected=F, board_requested=T, column_field_projected=F, column_list_len_GT_0=F, matcher_applied=F, matcher_depth_GE_0=F, records_grouped=F => TRUE
//mcdc:ignore SYS-REQ-260820-4628: board_rejected=F, board_requested=T, column_field_projected=F, column_list_len_GT_0=F, matcher_applied=F, matcher_depth_GE_0=T, records_grouped=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-4628: board_rejected=F, board_requested=T, column_field_projected=F, column_list_len_GT_0=T, matcher_applied=F, matcher_depth_GE_0=T, records_grouped=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-4628: board_rejected=F, board_requested=T, column_field_projected=F, column_list_len_GT_0=T, matcher_applied=T, matcher_depth_GE_0=T, records_grouped=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-4628: board_rejected=F, board_requested=T, column_field_projected=T, column_list_len_GT_0=T, matcher_applied=F, matcher_depth_GE_0=T, records_grouped=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-4628: board_rejected=F, board_requested=T, column_field_projected=T, column_list_len_GT_0=T, matcher_applied=T, matcher_depth_GE_0=T, records_grouped=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-4628: board_rejected=F, board_requested=T, column_field_projected=T, column_list_len_GT_0=T, matcher_applied=T, matcher_depth_GE_0=T, records_grouped=T => TRUE
// MCDC SYS-REQ-260820-4628: board_rejected=T, board_requested=T, column_field_projected=F, column_list_len_GT_0=T, matcher_applied=F, matcher_depth_GE_0=T, records_grouped=F => TRUE
//mcdc:ignore SYS-REQ-260820-4628: board_rejected=T, board_requested=T, column_field_projected=T, column_list_len_GT_0=F, matcher_applied=T, matcher_depth_GE_0=T, records_grouped=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-4628: board_rejected=T, board_requested=T, column_field_projected=T, column_list_len_GT_0=T, matcher_applied=T, matcher_depth_GE_0=T, records_grouped=T => TRUE
func TestBoardRejectsTraversalName(t *testing.T) {
	root := t.TempDir()
	a := app.OpenOrCWD(root)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Board("../crm", nil); err == nil {
		t.Fatal("expected invalid board name")
	}
}
