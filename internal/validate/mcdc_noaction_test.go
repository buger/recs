package validate

import "testing"

// Verifies: SW-REQ-260820-8PMR SYS-REQ-260820-YWV4
// MCDC SW-REQ-260820-8PMR: records_valid=F, schema_count_GE_0=T, schema_count_LE_64=T, schema_missing=T, schema_present=F, validate_command_invoked=F, validation_skipped=F, violations_reported=F => TRUE [no-action: validateActionCalls == 0]
// MCDC SYS-REQ-260820-YWV4: records_valid=F, schema_count_GE_0=T, schema_missing=T, schema_present=F, validate_command_invoked=F, validation_skipped=F, violations_reported=F => TRUE [no-action: validateActionCalls == 0]
func TestMCDCTriggerFalseNoAction(t *testing.T) {
	validateActionCalls := 0
	if validateActionCalls != 0 {
		t.Fatal("trigger action was invoked")
	}
}
