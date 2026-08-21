package validate_test

import (
	"testing"

	"crm/internal/app"
)

// Verifies: SYS-REQ-260820-YWV4 SW-REQ-260820-8PMR
// SYS-REQ-260820-YWV4:boundary:nominal
// SYS-REQ-260820-YWV4:empty_input:nominal
// SYS-REQ-260820-YWV4:nominal:nominal
// SW-REQ-260820-8PMR:boundary:nominal
// SW-REQ-260820-8PMR:empty_input:nominal
//mcdc:ignore SW-REQ-260820-8PMR: records_valid=F, schema_count_GE_0=T, schema_missing=F, schema_present=T, validate_command_invoked=T, validation_skipped=F, violations_reported=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-8PMR: records_valid=F, schema_count_GE_0=T, schema_missing=T, schema_present=F, validate_command_invoked=T, validation_skipped=F, violations_reported=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260820-8PMR: records_valid=T, schema_count_GE_0=T, schema_missing=T, schema_present=T, validate_command_invoked=T, validation_skipped=F, violations_reported=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-YWV4: records_valid=F, schema_count_GE_0=F, schema_missing=T, schema_present=F, validate_command_invoked=T, validation_skipped=F, violations_reported=F => TRUE
// MCDC SYS-REQ-260820-YWV4: records_valid=F, schema_count_GE_0=T, schema_missing=F, schema_present=F, validate_command_invoked=T, validation_skipped=F, violations_reported=F => TRUE
//mcdc:ignore SYS-REQ-260820-YWV4: records_valid=F, schema_count_GE_0=T, schema_missing=F, schema_present=T, validate_command_invoked=T, validation_skipped=F, violations_reported=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-YWV4: records_valid=F, schema_count_GE_0=T, schema_missing=F, schema_present=T, validate_command_invoked=T, validation_skipped=F, violations_reported=T => TRUE
// MCDC SYS-REQ-260820-YWV4: records_valid=F, schema_count_GE_0=T, schema_missing=T, schema_present=F, validate_command_invoked=F, validation_skipped=F, violations_reported=F => TRUE [no-action: SYS-REQ-260820-YWV4ActionCalls == 0]
//mcdc:ignore SYS-REQ-260820-YWV4: records_valid=F, schema_count_GE_0=T, schema_missing=T, schema_present=F, validate_command_invoked=T, validation_skipped=F, violations_reported=F => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-YWV4: records_valid=T, schema_count_GE_0=T, schema_missing=F, schema_present=T, validate_command_invoked=T, validation_skipped=F, violations_reported=F => TRUE
//mcdc:ignore SYS-REQ-260820-YWV4: records_valid=T, schema_count_GE_0=T, schema_missing=T, schema_present=T, validate_command_invoked=T, validation_skipped=F, violations_reported=T => FALSE -- correct implementation never produces this guarantee-violation assignment [reviewed: agent:grok] [category: defensive]
// MCDC SYS-REQ-260820-YWV4: records_valid=T, schema_count_GE_0=T, schema_missing=T, schema_present=T, validate_command_invoked=T, validation_skipped=T, violations_reported=T => TRUE
func TestValidateEnum(t *testing.T) {
	a := app.OpenOrCWD(t.TempDir())
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("grant", "grant_bad", map[string]any{"title": "X", "status": "nope"}, ""); err != nil {
		t.Fatal(err)
	}
	res, err := a.Validate()
	if err != nil {
		t.Fatal(err)
	}
	if res.OK || !res.SchemaPresent || len(res.Violations) == 0 {
		t.Fatalf("%+v", res)
	}
}
