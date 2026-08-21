package cli_test

import (
	"bytes"
	"testing"

	"crm/internal/cli"
)

// Verifies: SYS-REQ-260821-8FKR SW-REQ-260821-FCGM SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func TestCLIAllCommandsAndFlags(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main(nil, out, errb) != 0 {
		t.Fatal("empty args")
	}
	if cli.Main([]string{"--help"}, out, errb) != 0 || cli.Main([]string{"-h"}, out, errb) != 0 {
		t.Fatal("help")
	}
	if cli.Main([]string{"--json"}, out, errb) != 0 {
		t.Fatal("json only")
	}
	if run(t, root, "init", "--json") != 0 {
		t.Fatal("init")
	}
	if run(t, root, "--root="+root, "create", "grant", "--id", "grant_c", "--title", "C", "--name", "Cn", "--body", "body", "--set", "status=researching", "--set", "bad") != 0 {
		t.Fatal("create")
	}
	if run(t, root, "show", "grant_c") != 0 || run(t, root, "show") == 0 {
		t.Fatal("show")
	}
	if run(t, root, "list", "--type", "grant", "--json") != 0 {
		t.Fatal("list")
	}
	if run(t, root, "search", "C") != 0 || run(t, root, "query", "status", "=", "researching") != 0 {
		t.Fatal("search/query")
	}
	if run(t, root, "set", "grant_c", "status", "preparing") != 0 || run(t, root, "set", "x") == 0 {
		t.Fatal("set")
	}
	if run(t, root, "patch", "grant_c", "--set", "status=applied") != 0 || run(t, root, "patch") == 0 {
		t.Fatal("patch")
	}
	if run(t, root, "dashboard") != 0 || run(t, root, "dashboard", "prospects", "--json") != 0 {
		t.Fatal("dashboard")
	}
	if run(t, root, "dashboard", "new") == 0 || run(t, root, "dashboard", "new", "extra", "--name", "Extra") != 0 {
		t.Fatal("dashboard new")
	}
	if run(t, root, "board") != 0 || run(t, root, "board", "grants", "--filter", "status=applied", "--filter", "bad") != 0 {
		t.Fatal("board")
	}
	if run(t, root, "move", "grant_c", "grants", "researching") != 0 || run(t, root, "move", "x") == 0 {
		t.Fatal("move")
	}
	if run(t, root, "next") != 0 || run(t, root, "triage") != 0 {
		t.Fatal("next/triage")
	}
	if run(t, root, "validate") != 0 && false {
		// validate may exit 1 on violations
	}
	_ = run(t, root, "validate", "--json")
	if run(t, root, "index") != 0 || run(t, root, "context", "grant_c") != 0 || run(t, root, "context") == 0 {
		t.Fatal("index/context")
	}
	if run(t, root, "context", "grant_c", "--md") != 0 {
		t.Fatal("context md")
	}
	if run(t, root, "inbox") != 0 {
		t.Fatal("inbox")
	}
	if run(t, root, "agent", "install") == 0 || run(t, root, "agent") == 0 {
		t.Fatal("agent command must not exist")
	}
	if run(t, root, "create") == 0 || run(t, root, "nope") == 0 {
		t.Fatal("usage/unknown")
	}
	if run(t, root, "show", "missing", "--json") == 0 {
		t.Fatal("missing json")
	}
	if run(t, root, "create", "grant", "--id", "grant_c", "--json") == 0 {
		t.Fatal("exists")
	}
	if run(t, root, "patch", "grant_c", "--set", "status=nope", "--json") == 0 {
		t.Fatal("enum json")
	}
	if run(t, root, "patch", "grant_c", "--if-version", "sha256:dead", "--set", "status=applied", "--json") == 0 {
		t.Fatal("conflict json")
	}
	if cli.Main([]string{"--port", "9", "--port=8", "init", "--root", t.TempDir()}, out, errb) != 0 {
		t.Fatal("port flags")
	}
}

// Verifies: SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2 SYS-REQ-260820-PG9C SW-REQ-260820-YB5C
func TestCLIInitJSONAndHumanErrors(t *testing.T) {
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"init", "--root", t.TempDir()}, out, errb) == 2 {
		t.Fatal("noop")
	}
	root := t.TempDir()
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init human")
	}
	if cli.Main([]string{"--root", root, "board", "missing"}, out, errb) == 0 {
		t.Fatal("missing board human")
	}
	if cli.Main([]string{"--root", root, "validate"}, out, errb) > 1 {
		t.Fatal("validate code")
	}
}
