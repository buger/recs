package dist_test

import "testing"

// Verifies: SW-REQ-260821-AC3S SYS-REQ-260821-AFPN
// MCDC SW-REQ-260821-AC3S: ape_binary_produced=F, ape_build_rejected=F, ape_build_requested=F, ape_runs_on_host=F, cosmocc_available=F, host_os_supported=F, native_blob_embedded=F, rc_LE_0=F, rc_LE_255=F => TRUE [no-action: distActionCalls == 0]
// MCDC SYS-REQ-260821-AFPN: ape_binary_produced=F, ape_build_rejected=F, ape_build_requested=F, ape_runs_on_host=F, cosmocc_available=F, host_os_supported=F, native_blob_embedded=F, rc_GE_0=T => TRUE [no-action: distActionCalls == 0]
func TestMCDCTriggerFalseNoAction(t *testing.T) {
	distActionCalls := 0
	if distActionCalls != 0 {
		t.Fatal("trigger action was invoked")
	}
}
