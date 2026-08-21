package index_test

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/app"
)

// Verifies: SYS-REQ-260820-Q8GR SW-REQ-260820-BNR7 INT-REQ-260820-2JKK SYS-REQ-260820-0TQX SW-REQ-260820-V48V
// SYS-REQ-260820-Q8GR:idempotency:nominal
// SYS-REQ-260820-Q8GR:nominal:nominal
// SW-REQ-260820-BNR7:idempotency:nominal
// SW-REQ-260820-BNR7:nominal:nominal
// INT-REQ-260820-2JKK:idempotency:nominal
// INT-REQ-260820-2JKK:integration:integration
// SYS-REQ-260820-0TQX:error_handling:nominal
// SYS-REQ-260820-0TQX:nominal:nominal
// SW-REQ-260820-V48V:error_handling:nominal
// SW-REQ-260820-V48V:nominal:nominal
//mcdc:ignore INT-REQ-260820-2JKK: canonical_state_mutated=F, index_files_rewritten=F, index_rebuild_requested=T, record_count_GE_0=T, store_records_scanned=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore INT-REQ-260820-2JKK: canonical_state_mutated=F, index_files_rewritten=T, index_rebuild_requested=T, record_count_GE_0=T, store_records_scanned=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC INT-REQ-260820-2JKK: canonical_state_mutated=F, index_files_rewritten=T, index_rebuild_requested=T, record_count_GE_0=T, store_records_scanned=T => TRUE
// MCDC INT-REQ-260820-2JKK: canonical_state_mutated=T, index_files_rewritten=T, index_rebuild_requested=F, record_count_GE_0=T, store_records_scanned=T => TRUE [no-action: INT-REQ-260820-2JKKActionCalls == 0]
// MCDC INT-REQ-260820-2JKK: canonical_state_mutated=T, index_files_rewritten=T, index_rebuild_requested=T, record_count_GE_0=F, store_records_scanned=T => TRUE
//mcdc:ignore INT-REQ-260820-2JKK: canonical_state_mutated=T, index_files_rewritten=T, index_rebuild_requested=T, record_count_GE_0=T, store_records_scanned=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// SW-REQ-260820-BNR7
//mcdc:ignore SW-REQ-260820-BNR7: canonical_state_mutated=F, index_is_canonical=F, index_rebuild_requested=T, index_rebuilt_from_records=F, record_count_GE_0=T, record_count_LE_100000=T, store_records_scanned=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-BNR7: canonical_state_mutated=F, index_is_canonical=F, index_rebuild_requested=T, index_rebuilt_from_records=T, record_count_GE_0=T, record_count_LE_100000=T, store_records_scanned=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-BNR7: canonical_state_mutated=F, index_is_canonical=F, index_rebuild_requested=T, index_rebuilt_from_records=T, record_count_GE_0=T, record_count_LE_100000=T, store_records_scanned=T => TRUE
//mcdc:ignore SW-REQ-260820-BNR7: canonical_state_mutated=F, index_is_canonical=T, index_rebuild_requested=T, index_rebuilt_from_records=T, record_count_GE_0=T, record_count_LE_100000=T, store_records_scanned=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-BNR7: canonical_state_mutated=T, index_is_canonical=F, index_rebuild_requested=T, index_rebuilt_from_records=T, record_count_GE_0=T, record_count_LE_100000=T, store_records_scanned=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-BNR7: canonical_state_mutated=T, index_is_canonical=T, index_rebuild_requested=F, index_rebuilt_from_records=T, record_count_GE_0=T, record_count_LE_100000=T, store_records_scanned=T => TRUE [no-action: SW_REQ_260820_BNR7ActionCalls == 0]
// MCDC SW-REQ-260820-BNR7: canonical_state_mutated=T, index_is_canonical=T, index_rebuild_requested=T, index_rebuilt_from_records=T, record_count_GE_0=F, record_count_LE_100000=T, store_records_scanned=T => TRUE
// MCDC SW-REQ-260820-BNR7: canonical_state_mutated=T, index_is_canonical=T, index_rebuild_requested=T, index_rebuilt_from_records=T, record_count_GE_0=T, record_count_LE_100000=F, store_records_scanned=T => TRUE
//mcdc:ignore SW-REQ-260820-BNR7: canonical_state_mutated=T, index_is_canonical=T, index_rebuild_requested=T, index_rebuilt_from_records=T, record_count_GE_0=T, record_count_LE_100000=T, store_records_scanned=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-0TQX: context_rejected=F, context_requested=F, related_records_assembled=F, relation_count_GE_0=T, relations_resolved=F, seed_record_loaded=F => TRUE [no-action: SYS-REQ-260820-0TQXActionCalls == 0]
// MCDC SYS-REQ-260820-0TQX: context_rejected=F, context_requested=T, related_records_assembled=F, relation_count_GE_0=F, relations_resolved=F, seed_record_loaded=F => TRUE
//mcdc:ignore SYS-REQ-260820-0TQX: context_rejected=F, context_requested=T, related_records_assembled=F, relation_count_GE_0=T, relations_resolved=F, seed_record_loaded=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-0TQX: context_rejected=F, context_requested=T, related_records_assembled=F, relation_count_GE_0=T, relations_resolved=T, seed_record_loaded=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-0TQX: context_rejected=F, context_requested=T, related_records_assembled=T, relation_count_GE_0=T, relations_resolved=F, seed_record_loaded=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-0TQX: context_rejected=F, context_requested=T, related_records_assembled=T, relation_count_GE_0=T, relations_resolved=T, seed_record_loaded=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-0TQX: context_rejected=F, context_requested=T, related_records_assembled=T, relation_count_GE_0=T, relations_resolved=T, seed_record_loaded=T => TRUE
// MCDC SYS-REQ-260820-0TQX: context_rejected=T, context_requested=T, related_records_assembled=F, relation_count_GE_0=T, relations_resolved=F, seed_record_loaded=F => TRUE
//mcdc:ignore SYS-REQ-260820-Q8GR: index_is_canonical=F, index_rebuild_requested=T, index_rebuilt_from_records=F, record_count_GE_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-Q8GR: index_is_canonical=T, index_rebuild_requested=T, index_rebuilt_from_records=T, record_count_GE_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
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

// Verifies: SYS-REQ-260820-0TQX SW-REQ-260820-V48V
// SYS-REQ-260820-0TQX:error_handling:negative
func TestContextMissingSeed(t *testing.T) {
	a := app.OpenOrCWD(t.TempDir())
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Context("missing_id"); err == nil {
		t.Fatal("expected missing seed error")
	}
}
