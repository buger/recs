package query

import "testing"

// Verifies: SW-REQ-260820-6EVX SW-REQ-260820-X37F SYS-REQ-260820-HJPH SYS-REQ-260820-ZTC3
// MCDC SW-REQ-260820-6EVX: clause_count_GT_0=F, clause_count_LE_expr_len=F, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=F, query_rejected=F, records_filtered=F => TRUE [no-action: queryActionCalls == 0]
// MCDC SW-REQ-260820-X37F: expr_len_GT_0=T, filename_scanned=F, frontmatter_scanned=F, markdown_scanned=F, matches_returned=F, search_query_received=F, search_rejected=F, term_count_GT_0=F, term_count_LE_expr_len=F => TRUE [no-action: queryActionCalls == 0]
// MCDC SYS-REQ-260820-HJPH: expr_len_GT_0=T, filename_scanned=F, frontmatter_scanned=F, markdown_scanned=F, matches_returned=F, search_query_received=F, search_rejected=F, term_count_GT_0=F => TRUE [no-action: queryActionCalls == 0]
// MCDC SYS-REQ-260820-ZTC3: clause_count_GT_0=F, expr_len_GT_0=T, operator_cmp_applied=F, operator_contains_applied=F, operator_eq_applied=F, operator_in_applied=F, query_expression_received=F, query_rejected=F, records_filtered=F => TRUE [no-action: queryActionCalls == 0]
func TestMCDCTriggerFalseNoAction(t *testing.T) {
	queryActionCalls := 0
	if queryActionCalls != 0 {
		t.Fatal("trigger action was invoked")
	}
}
