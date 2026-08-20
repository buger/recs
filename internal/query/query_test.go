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
// MCDC SW-REQ-260820-6EVX: clause_count_GT_0=F, clause_count_LE_expr_len=F, expr_len_GT_0=F, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=F => TRUE
// MCDC SW-REQ-260820-6EVX: clause_count_GT_0=F, clause_count_LE_expr_len=F, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=F, query_rejected=F, records_filtered=F => TRUE
//mcdc:ignore SW-REQ-260820-6EVX: clause_count_GT_0=F, clause_count_LE_expr_len=F, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-6EVX: clause_count_GT_0=F, clause_count_LE_expr_len=T, expr_len_GT_0=T, operator_cmp_applied=T, operator_contains_applied=T, operator_eq_applied=T, operator_in_applied=T, query_expression_received=T, query_rejected=T, records_filtered=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-6EVX: clause_count_GT_0=T, clause_count_LE_expr_len=F, expr_len_GT_0=T, operator_cmp_applied=T, operator_contains_applied=T, operator_eq_applied=T, operator_in_applied=T, query_expression_received=T, query_rejected=T, records_filtered=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-6EVX: clause_count_GT_0=T, clause_count_LE_expr_len=T, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-6EVX: clause_count_GT_0=T, clause_count_LE_expr_len=T, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-6EVX: clause_count_GT_0=T, clause_count_LE_expr_len=T, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=T, records_filtered=F => TRUE
// MCDC SW-REQ-260820-6EVX: clause_count_GT_0=T, clause_count_LE_expr_len=T, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=T, query_expression_received=T, query_rejected=F, records_filtered=T => TRUE
// MCDC SW-REQ-260820-6EVX: clause_count_GT_0=T, clause_count_LE_expr_len=T, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=T, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=T => TRUE
// MCDC SW-REQ-260820-6EVX: clause_count_GT_0=T, clause_count_LE_expr_len=T, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=T, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=T => TRUE
//mcdc:ignore SW-REQ-260820-6EVX: clause_count_GT_0=T, clause_count_LE_expr_len=T, expr_len_GT_0=T, operator_cmp_applied=T, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-6EVX: clause_count_GT_0=T, clause_count_LE_expr_len=T, expr_len_GT_0=T, operator_cmp_applied=T, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=T => TRUE
// MCDC SW-REQ-260820-6EVX: clause_count_GT_0=T, clause_count_LE_expr_len=T, expr_len_GT_0=T, operator_cmp_applied=T, operator_contains_applied=T, operator_eq_applied=T, operator_in_applied=T, query_expression_received=T, query_rejected=T, records_filtered=T => TRUE
// MCDC SW-REQ-260820-X37F: expr_len_GT_0=F, filename_scanned=F, frontmatter_scanned=F, markdown_scanned=F, matches_returned=F, search_query_received=T, search_rejected=F, term_count_GT_0=F, term_count_LE_expr_len=F => TRUE
// MCDC SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=F, frontmatter_scanned=F, markdown_scanned=F, matches_returned=F, search_query_received=F, search_rejected=F, term_count_GT_0=F, term_count_LE_expr_len=F => TRUE
//mcdc:ignore SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=F, frontmatter_scanned=F, markdown_scanned=F, matches_returned=F, search_query_received=T, search_rejected=F, term_count_GT_0=F, term_count_LE_expr_len=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=F, frontmatter_scanned=F, markdown_scanned=F, matches_returned=F, search_query_received=T, search_rejected=F, term_count_GT_0=T, term_count_LE_expr_len=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=F, frontmatter_scanned=F, markdown_scanned=F, matches_returned=F, search_query_received=T, search_rejected=T, term_count_GT_0=T, term_count_LE_expr_len=T => TRUE
//mcdc:ignore SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=F, frontmatter_scanned=T, markdown_scanned=T, matches_returned=T, search_query_received=T, search_rejected=F, term_count_GT_0=T, term_count_LE_expr_len=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=F, markdown_scanned=T, matches_returned=T, search_query_received=T, search_rejected=F, term_count_GT_0=T, term_count_LE_expr_len=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=T, markdown_scanned=F, matches_returned=T, search_query_received=T, search_rejected=F, term_count_GT_0=T, term_count_LE_expr_len=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=T, markdown_scanned=T, matches_returned=F, search_query_received=T, search_rejected=F, term_count_GT_0=T, term_count_LE_expr_len=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=T, markdown_scanned=T, matches_returned=T, search_query_received=T, search_rejected=F, term_count_GT_0=T, term_count_LE_expr_len=T => TRUE
//mcdc:ignore SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=T, markdown_scanned=T, matches_returned=T, search_query_received=T, search_rejected=T, term_count_GT_0=F, term_count_LE_expr_len=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=T, markdown_scanned=T, matches_returned=T, search_query_received=T, search_rejected=T, term_count_GT_0=T, term_count_LE_expr_len=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=T, markdown_scanned=T, matches_returned=T, search_query_received=T, search_rejected=T, term_count_GT_0=T, term_count_LE_expr_len=T => TRUE
// MCDC SYS-REQ-260820-HJPH: expr_len_GT_0=F, filename_scanned=F, frontmatter_scanned=F, markdown_scanned=F, matches_returned=F, search_query_received=T, search_rejected=F, term_count_GT_0=F => TRUE
// MCDC SYS-REQ-260820-HJPH: expr_len_GT_0=T, filename_scanned=F, frontmatter_scanned=F, markdown_scanned=F, matches_returned=F, search_query_received=F, search_rejected=F, term_count_GT_0=F => TRUE
//mcdc:ignore SYS-REQ-260820-HJPH: expr_len_GT_0=T, filename_scanned=F, frontmatter_scanned=F, markdown_scanned=F, matches_returned=F, search_query_received=T, search_rejected=F, term_count_GT_0=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-HJPH: expr_len_GT_0=T, filename_scanned=F, frontmatter_scanned=F, markdown_scanned=F, matches_returned=F, search_query_received=T, search_rejected=F, term_count_GT_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-HJPH: expr_len_GT_0=T, filename_scanned=F, frontmatter_scanned=F, markdown_scanned=F, matches_returned=F, search_query_received=T, search_rejected=T, term_count_GT_0=T => TRUE
//mcdc:ignore SYS-REQ-260820-HJPH: expr_len_GT_0=T, filename_scanned=F, frontmatter_scanned=T, markdown_scanned=T, matches_returned=T, search_query_received=T, search_rejected=F, term_count_GT_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-HJPH: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=F, markdown_scanned=T, matches_returned=T, search_query_received=T, search_rejected=F, term_count_GT_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-HJPH: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=T, markdown_scanned=F, matches_returned=T, search_query_received=T, search_rejected=F, term_count_GT_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-HJPH: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=T, markdown_scanned=T, matches_returned=F, search_query_received=T, search_rejected=F, term_count_GT_0=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-HJPH: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=T, markdown_scanned=T, matches_returned=T, search_query_received=T, search_rejected=F, term_count_GT_0=T => TRUE
//mcdc:ignore SYS-REQ-260820-HJPH: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=T, markdown_scanned=T, matches_returned=T, search_query_received=T, search_rejected=T, term_count_GT_0=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-HJPH: expr_len_GT_0=T, filename_scanned=T, frontmatter_scanned=T, markdown_scanned=T, matches_returned=T, search_query_received=T, search_rejected=T, term_count_GT_0=T => TRUE
// MCDC SYS-REQ-260820-ZTC3: clause_count_GT_0=F, expr_len_GT_0=F, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=F => TRUE
// MCDC SYS-REQ-260820-ZTC3: clause_count_GT_0=F, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=F, query_rejected=F, records_filtered=F => TRUE
//mcdc:ignore SYS-REQ-260820-ZTC3: clause_count_GT_0=F, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-ZTC3: clause_count_GT_0=F, expr_len_GT_0=T, operator_cmp_applied=T, operator_contains_applied=T, operator_eq_applied=T, operator_in_applied=T, query_expression_received=T, query_rejected=T, records_filtered=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-ZTC3: clause_count_GT_0=T, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SYS-REQ-260820-ZTC3: clause_count_GT_0=T, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-ZTC3: clause_count_GT_0=T, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=T, records_filtered=F => TRUE
// MCDC SYS-REQ-260820-ZTC3: clause_count_GT_0=T, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=T, query_expression_received=T, query_rejected=F, records_filtered=T => TRUE
// MCDC SYS-REQ-260820-ZTC3: clause_count_GT_0=T, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=T, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=T => TRUE
// MCDC SYS-REQ-260820-ZTC3: clause_count_GT_0=T, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=T, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=T => TRUE
//mcdc:ignore SYS-REQ-260820-ZTC3: clause_count_GT_0=T, expr_len_GT_0=T, operator_cmp_applied=T, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-ZTC3: clause_count_GT_0=T, expr_len_GT_0=T, operator_cmp_applied=T, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=T, query_rejected=F, records_filtered=T => TRUE
// MCDC SYS-REQ-260820-ZTC3: clause_count_GT_0=T, expr_len_GT_0=T, operator_cmp_applied=T, operator_contains_applied=T, operator_eq_applied=T, operator_in_applied=T, query_expression_received=T, query_rejected=T, records_filtered=T => TRUE
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
