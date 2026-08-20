package contextpkg_test

import (
	"testing"

	"crm/internal/contextpkg"
	"crm/internal/record"
)

// Verifies: SW-REQ-260820-V48V SYS-REQ-260820-0TQX
// SW-REQ-260820-V48V:error_handling:negative
func TestAssembleIgnoresUnknownIDs(t *testing.T) {
	seed := &record.Record{ID: "person_a", Type: "person", Fields: map[string]any{"id": "person_a", "company": "missing_co"}}
	got := contextpkg.Assemble(seed, []*record.Record{seed})
	if got == nil || got.Seed != seed {
		t.Fatal("seed")
	}
	if len(got.Related) != 0 {
		t.Fatalf("unknown ids should not appear: %#v", got.Related)
	}
}
