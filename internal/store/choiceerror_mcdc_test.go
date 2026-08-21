package store

import "testing"

// Verifies: SW-REQ-260821-AY8F SW-REQ-260821-E5V8
func TestChoiceErrorMsgIndependence(t *testing.T) {
	with := &ChoiceError{Code: "unknown_flag", Msg: "unknown flag --x"}
	if with.Error() != "unknown flag --x" {
		t.Fatalf("msg: %q", with.Error())
	}
	empty := &ChoiceError{Code: "unknown_flag"}
	if empty.Error() != "unknown_flag" {
		t.Fatalf("code fallback: %q", empty.Error())
	}
}
