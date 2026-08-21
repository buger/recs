package contextpkg_test

import (
	"testing"

	"github.com/buger/recs/internal/contextpkg"
	"github.com/buger/recs/internal/record"
)

// Verifies: SW-REQ-260820-V48V SYS-REQ-260820-0TQX
// SW-REQ-260820-V48V:error_handling:negative
//mcdc:ignore SW-REQ-260820-V48V: context_rejected=F, context_requested=T, related_records_assembled=F, relation_count_GE_0=T, relations_resolved=F, seed_record_loaded=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-V48V: context_rejected=F, context_requested=T, related_records_assembled=F, relation_count_GE_0=T, relations_resolved=T, seed_record_loaded=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-V48V: context_rejected=F, context_requested=T, related_records_assembled=T, relation_count_GE_0=T, relations_resolved=F, seed_record_loaded=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-V48V: context_rejected=F, context_requested=T, related_records_assembled=T, relation_count_GE_0=T, relations_resolved=T, seed_record_loaded=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
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
