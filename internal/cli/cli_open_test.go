package cli

import (
	"testing"
)

// Verifies: SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2
func TestOpenAppInitBranch(t *testing.T) {
	a, err := openApp("init", t.TempDir())
	if err != nil || a == nil {
		t.Fatal(err, a)
	}
	if _, err := openApp("list", t.TempDir()); err == nil {
		t.Fatal("expected missing workspace")
	}
}
