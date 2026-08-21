package store_test

import (
	"errors"
	"sync"
	"testing"

	"crm/internal/store"
)

// Verifies: SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4 INT-REQ-260821-5BJJ
// SYS-REQ-260820-2SQZ:concurrent:race
// SW-REQ-260820-Q3C4:concurrent:race
// INT-REQ-260821-5BJJ:concurrent:race
// SYS-REQ-260820-2SQZ:concurrent_invariant_preserved:race
// SW-REQ-260820-Q3C4:concurrent_invariant_preserved:race
// SYS-REQ-260820-2SQZ:concurrent:nominal
// SW-REQ-260820-Q3C4:concurrent:nominal
// INT-REQ-260821-5BJJ:concurrent:nominal
// INT-REQ-260821-5BJJ:integration:integration
// MCDC INT-REQ-260821-5BJJ: conflict_reported=F, parse_budget_GE_0=T, record_mutation_requested=T, version_mismatch=F => TRUE
// MCDC INT-REQ-260821-5BJJ: conflict_reported=T, parse_budget_GE_0=T, record_mutation_requested=T, version_mismatch=T => TRUE
//mcdc:ignore INT-REQ-260821-5BJJ: conflict_reported=F, parse_budget_GE_0=T, record_mutation_requested=T, version_mismatch=T => FALSE -- version mismatch without conflict is the literal negation of the serialize-or-conflict contract [reviewed: agent:grok] [category: defensive]
func TestConcurrentPatchesSerializeOrConflict(t *testing.T) {
	a := initApp(t)
	rec, err := a.Create("grant", "grant_race", map[string]any{"title": "Race", "status": "researching"}, "")
	if err != nil {
		t.Fatal(err)
	}
	version := rec.Version()

	var wg sync.WaitGroup
	type result struct {
		err error
	}
	out := make(chan result, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := a.Patch(rec.ID, map[string]any{"status": "applied"}, nil, version)
		out <- result{err: err}
	}()
	go func() {
		defer wg.Done()
		_, err := a.Patch(rec.ID, map[string]any{"status": "rejected"}, nil, version)
		out <- result{err: err}
	}()
	wg.Wait()
	close(out)

	var ok, conflict, other int
	for r := range out {
		if r.err == nil {
			ok++
			continue
		}
		var conf *store.ConflictError
		if errors.As(r.err, &conf) {
			conflict++
			continue
		}
		other++
		t.Errorf("unexpected patch error: %v", r.err)
	}
	if ok == 0 {
		t.Fatal("expected at least one concurrent patch to commit")
	}
	if ok+conflict != 2 || other != 0 {
		t.Fatalf("expected serial success or conflict, got ok=%d conflict=%d other=%d", ok, conflict, other)
	}
	got, err := a.Show(rec.ID)
	if err != nil {
		t.Fatal(err)
	}
	status := got.GetString("status")
	if status != "applied" && status != "rejected" {
		t.Fatalf("torn or lost status %q", status)
	}
	if got.ID != rec.ID || got.Path != rec.Path {
		t.Fatalf("concurrent patch moved or retargeted record: id=%s path=%s", got.ID, got.Path)
	}
}
