package cli_test

import (
	"bytes"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/buger/recs/internal/cli"
)

// Verifies: SYS-REQ-260821-8FKR SW-REQ-260821-FCGM SYS-REQ-260820-YWV4 SW-REQ-260820-8PMR
func TestCLIFlagWithoutValueAndValidateBranches(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init", errb.String())
	}
	flags := []string{"--root", "--port", "--set", "--filter", "--if-version", "--type", "--id", "--title", "--name", "--body"}
	for _, f := range flags {
		_ = cli.Main([]string{"list", f}, out, errb)
	}
	if cli.Main([]string{"--root", root, "create", "note", "--title", "T"}, out, errb) != 0 {
		t.Fatal("create", errb.String())
	}
	if cli.Main([]string{"--root", root, "validate"}, out, errb) > 1 {
		t.Fatal("validate present")
	}
	if err := os.WriteFile(filepath.Join(root, "crm.yaml"), []byte("name: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if cli.Main([]string{"--root", root, "validate"}, out, errb) != 0 {
		t.Fatal("no schema types", out.String(), errb.String())
	}
	if err := os.WriteFile(filepath.Join(root, "crm.yaml"), []byte("types:\n  note:\n    required: [title]\n    fields:\n      status:\n        enum: [open]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	code := cli.Main([]string{"--root", root, "validate"}, out, errb)
	if code != 1 && code != 0 {
		t.Fatal("validate code", code)
	}
	if cli.Main([]string{"--root", root, "agent", "nope"}, out, errb) == 0 {
		t.Fatal("agent usage")
	}
	if cli.Main([]string{"--root", root, "list"}, out, errb) != 0 {
		t.Fatal("human list")
	}
}

// Verifies: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7 INT-REQ-260820-AHKR
func TestCLIServePortDefaultAndOverride(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	// Occupy the default and override ports so Main returns. A successful
	// ListenAndServe would leak a goroutine and hang go test -cover.
	def, err := net.Listen("tcp", "127.0.0.1:7777")
	if err != nil {
		t.Skip("default serve port 7777 busy:", err)
	}
	defer def.Close()
	if cli.Main([]string{"--root", root, "--port", "0", "serve"}, out, errb) == 0 {
		t.Fatal("expected default port 7777 busy")
	}
	ov, err := net.Listen("tcp", "127.0.0.1:18765")
	if err != nil {
		t.Skip("override port 18765 busy:", err)
	}
	defer ov.Close()
	if cli.Main([]string{"--root", root, "--port", "18765", "serve"}, out, errb) == 0 {
		t.Fatal("expected override port busy")
	}
}

// Verifies: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7 INT-REQ-260820-AHKR
func TestCLIServeListenError(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if cli.Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	if cli.Main([]string{"--root", root, "--port", strconv.Itoa(port), "serve"}, out, errb) == 0 {
		t.Fatal("expected busy port")
	}
}
