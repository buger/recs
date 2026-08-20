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
