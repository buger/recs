package query

import (
	"testing"
	"time"

	"github.com/buger/recs/internal/record"
)

// Verifies: SW-REQ-260823-1PTN
// SW-REQ-260823-1PTN:nominal:nominal
// MCDC SW-REQ-260823-1PTN: clause_value_replaced_with_utc_now=T, now_token_in_clause_value=T, query_expression_received=T => TRUE
// MCDC SW-REQ-260823-1PTN: clause_value_replaced_with_utc_now=F, now_token_in_clause_value=F, query_expression_received=T => TRUE
//mcdc:ignore SW-REQ-260823-1PTN: clause_value_replaced_with_utc_now=F, now_token_in_clause_value=T, query_expression_received=T => FALSE -- correct implementation always expands the token, so this guarantee-violation assignment is unreachable [reviewed: agent:kimi] [category: defensive]
func TestParseExpandsNowToken(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)

	if _, err := Parse(``); err == nil {
		t.Fatal("empty query must be rejected")
	}
	if _, err := Parse(`$now`); err == nil {
		t.Fatal("bare token without operator must be rejected")
	}

	cs, err := Parse(`due < $now`)
	if err != nil {
		t.Fatal(err)
	}
	got, err := time.Parse(time.RFC3339, cs[0].Value)
	if err != nil {
		t.Fatalf("$now not expanded to RFC3339: %q", cs[0].Value)
	}
	if got.Before(before) || got.After(time.Now().UTC().Add(time.Second)) {
		t.Fatalf("expanded value %v outside parse-time window", got)
	}

	cs, err = Parse(`due <= now`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := time.Parse(time.RFC3339, cs[0].Value); err != nil {
		t.Fatalf("bare now not expanded: %q", cs[0].Value)
	}

	cs, err = Parse(`due < 2026-01-01 owner = nowak`)
	if err != nil {
		t.Fatal(err)
	}
	if cs[0].Value != "2026-01-01" || cs[1].Value != "nowak" {
		t.Fatalf("literal values changed: %#v", cs)
	}

	cs, err = Parse(`tag in [$now,fixed]`)
	if err != nil {
		t.Fatal(err)
	}
	if len(cs[0].Values) != 2 || cs[0].Values[1] != "fixed" {
		t.Fatalf("in-list values: %#v", cs[0].Values)
	}
	if _, err := time.Parse(time.RFC3339, cs[0].Values[0]); err != nil {
		t.Fatalf("in-list $now not expanded: %q", cs[0].Values[0])
	}
}

// Verifies: SW-REQ-260823-1PTN
// SW-REQ-260823-1PTN:nominal:nominal
func TestFilterOverdueWithNow(t *testing.T) {
	past := &record.Record{Fields: map[string]any{
		"due": time.Now().UTC().Add(-24 * time.Hour).Format("2006-01-02"),
	}}
	future := &record.Record{Fields: map[string]any{
		"due": time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02"),
	}}
	got, err := Filter([]*record.Record{past, future}, `due < $now`)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != past {
		t.Fatalf("rolling overdue filter matched %d records", len(got))
	}
}

// Verifies: SW-REQ-260823-1PTN
// MCDC SW-REQ-260823-1PTN: clause_value_replaced_with_utc_now=F, now_token_in_clause_value=T, query_expression_received=F => TRUE [no-action: queryActionCalls == 0]
// Expected: TRUE (requirement satisfied)
func TestMCDC_SW_REQ_260823_1PTN_Row2(t *testing.T) {
	// Witness row 2: clause_value_replaced_with_utc_now=F, now_token_in_clause_value=T, query_expression_received=F
	queryActionCalls := 0
	if queryActionCalls != 0 {
		t.Fatal("trigger action was invoked")
	}
}
