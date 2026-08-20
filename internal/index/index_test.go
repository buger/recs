package index_test

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/app"
)

// Verifies: SYS-REQ-260820-Q8GR SW-REQ-260820-BNR7 INT-REQ-260820-2JKK SYS-REQ-260820-0TQX SW-REQ-260820-V48V
func TestIndexAndContext(t *testing.T) {
	a := app.OpenOrCWD(t.TempDir())
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("company", "company_acme", map[string]any{"name": "Acme"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("person", "person_alice", map[string]any{"name": "Alice", "company": "company_acme"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.RebuildIndex(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(a.Root(), ".crm/index/records.json")); err != nil {
		t.Fatal(err)
	}
	_ = os.RemoveAll(filepath.Join(a.Root(), ".crm/index"))
	if _, err := a.RebuildIndex(); err != nil {
		t.Fatal(err)
	}
	bundle, err := a.Context("company_acme")
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Related) != 1 || bundle.Related[0].ID != "person_alice" {
		t.Fatalf("%+v", bundle.Related)
	}
}
