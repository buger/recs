package cli

import (
	"bytes"
	"net"
	"strings"
	"testing"
)

// Verifies: SYS-REQ-260821-8FKR SW-REQ-260821-FCGM SW-REQ-260821-E5V8 SW-REQ-260821-AY8F SW-REQ-260821-9737
func TestNextHintIndependence(t *testing.T) {
	// unknown flag: cmd non-empty and not help (T,T)
	if got := nextHint("list", "unknown flag --nope"); got != "recs help list" {
		t.Fatalf("unknown flag with cmd: %q", got)
	}
	// unknown flag: empty cmd (F, skip)
	if got := nextHint("", "unknown flag --nope"); got != "recs --help" {
		t.Fatalf("unknown flag empty cmd: %q", got)
	}
	// unknown flag: cmd is help (T,F)
	if got := nextHint("help", "unknown flag --nope"); got != "recs --help" {
		t.Fatalf("unknown flag help cmd: %q", got)
	}
	// dashboard + not found (T,T) — must be checked before the board substring
	if got := nextHint("show", "dashboard home not found"); got != "recs board" {
		t.Fatalf("dashboard not found: %q", got)
	}
	// dashboard without not found (T,F) falls through
	if got := nextHint("show", "dashboard layout invalid"); got == "recs board" {
		t.Fatalf("dashboard without not-found should not hint board, got %q", got)
	}
	// board + not found, no dashboard (board T,T after dashboard F)
	if got := nextHint("board", "board grants not found"); got != "recs board" {
		t.Fatalf("board not found: %q", got)
	}
}

// Verifies: SW-REQ-260821-9737 SW-REQ-260821-AY8F SW-REQ-260821-MFR2 SYS-REQ-260821-8FKR SW-REQ-260820-6EVX
func TestMainHotspotIndependence(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init", errb.String())
	}

	// cmd == "" when a flag errors before any command token
	if Main([]string{"--not-a-flag"}, out, errb) == 0 {
		t.Fatal("expected unknown flag with empty command")
	}

	// --set= form (HasPrefix true) and parseAssignment success
	if Main([]string{"--root", root, "create", "note", "--id", "n1", "--title", "N"}, out, errb) != 0 {
		t.Fatal("create", errb.String())
	}
	if Main([]string{"--root", root, "patch", "n1", "--set=status=open"}, out, errb) != 0 {
		t.Fatal("set= form", errb.String())
	}

	// parseAssignment: empty key with '=' present (--set==value)
	if Main([]string{"--root", root, "patch", "n1", "--set==value"}, out, errb) == 0 {
		t.Fatal("expected empty-key --set=")
	}
	// parseAssignment error on --set= without '='
	if Main([]string{"--root", root, "patch", "n1", "--set=nokeq"}, out, errb) == 0 {
		t.Fatal("expected malformed --set=")
	}
	// --filter assignment error
	if Main([]string{"--root", root, "list", "--filter", "nokeq"}, out, errb) == 0 {
		t.Fatal("expected malformed --filter")
	}

	// ingest single-token email / record (the || at len(rest)==1)
	if Main([]string{"--root", root, "ingest", "email"}, out, errb) == 0 {
		t.Fatal("ingest email without source should fail")
	}
	if Main([]string{"--root", root, "ingest", "record"}, out, errb) == 0 {
		t.Fatal("ingest record without source should fail")
	}

	// invalid query without unknown operator
	if Main([]string{"--root", root, "query", "notaclause", "--json"}, out, errb) == 0 {
		t.Fatal("expected invalid query")
	}
	if !strings.Contains(out.String()+errb.String(), "invalid query") && !strings.Contains(out.String(), "unknown_operator") {
		t.Log("invalid query payload", out.String(), errb.String())
	}

	// next with a real row so human != ""
	if Main([]string{"--root", root, "create", "task", "--id", "t1", "--title", "Due", "--set=status=open"}, out, errb) != 0 {
		t.Fatal("create task", errb.String())
	}
	out.Reset()
	errb.Reset()
	if Main([]string{"--root", root, "next"}, out, errb) != 0 {
		t.Fatal("next", errb.String())
	}
	if !strings.Contains(out.String(), "t1") && !strings.Contains(out.String(), "Due") {
		t.Fatalf("expected next row, got %q", out.String())
	}
}

// Verifies: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7
func TestServeDefaultPortEvaluatesZero(t *testing.T) {
	root := t.TempDir()
	out, errb := &bytes.Buffer{}, &bytes.Buffer{}
	if Main([]string{"--root", root, "init"}, out, errb) != 0 {
		t.Fatal("init")
	}
	ln, err := net.Listen("tcp", "127.0.0.1:7777")
	if err == nil {
		defer ln.Close()
	}
	// 7777 is ours or someone else's; serve without --port still evaluates port==0.
	if Main([]string{"--root", root, "serve"}, out, errb) == 0 {
		t.Fatal("expected default port bind to fail")
	}
}
