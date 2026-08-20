package query_test

import (
	"testing"

	"crm/internal/app"
)

// Verifies: SYS-REQ-260820-ZTC3 SW-REQ-260820-6EVX SYS-REQ-260820-HJPH SW-REQ-260820-X37F
// SYS-REQ-260820-ZTC3:determinism:nominal
// SYS-REQ-260820-ZTC3:malformed_input:nominal
// SW-REQ-260820-6EVX:determinism:nominal
// SW-REQ-260820-6EVX:malformed_input:nominal
// SYS-REQ-260820-HJPH:empty_input:nominal
// SYS-REQ-260820-HJPH:nominal:nominal
// SW-REQ-260820-X37F:empty_input:nominal
// SW-REQ-260820-X37F:nominal:nominal
func TestQueryAndSearch(t *testing.T) {
	root := t.TempDir()
	a := app.OpenOrCWD(root)
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("grant", "grant_a", map[string]any{"title": "Solana MC/DC", "status": "preparing", "tags": []any{"solana"}}, "formal verification"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("person", "person_alice", map[string]any{"name": "Alice", "status": "active"}, "intro"); err != nil {
		t.Fatal(err)
	}
	got, err := a.Query("type=grant status=preparing")
	if err != nil || len(got) != 1 || got[0].ID != "grant_a" {
		t.Fatalf("query: %v %#v", err, got)
	}
	got, err = a.Query("tags contains solana")
	if err != nil || len(got) != 1 {
		t.Fatalf("contains: %v %#v", err, got)
	}
	got, err = a.Query("status in preparing,applied")
	if err != nil || len(got) != 1 {
		t.Fatalf("in: %v %#v", err, got)
	}
	found, err := a.Search("MCDC Solana")
	if err != nil || len(found) != 1 {
		t.Fatalf("search: %v %#v", err, found)
	}
}

// Verifies: SYS-REQ-260820-ZTC3 SW-REQ-260820-6EVX
// SYS-REQ-260820-ZTC3:malformed_input:negative
// SW-REQ-260820-6EVX:malformed_input:negative
func TestQueryRejectsMalformed(t *testing.T) {
	a := app.OpenOrCWD(t.TempDir())
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Query(""); err == nil {
		t.Fatal("expected empty query error")
	}
	if _, err := a.Query("not-a-clause"); err == nil {
		t.Fatal("expected malformed query error")
	}
}

// Verifies: SYS-REQ-260820-HJPH SW-REQ-260820-X37F
// SYS-REQ-260820-HJPH:empty_input:negative
// SW-REQ-260820-X37F:empty_input:negative
func TestSearchEmptyReturnsNone(t *testing.T) {
	a := app.OpenOrCWD(t.TempDir())
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("grant", "grant_s", map[string]any{"title": "S", "status": "researching"}, ""); err != nil {
		t.Fatal(err)
	}
	got, err := a.Search("   ")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("empty search should match none: %#v", got)
	}
}
