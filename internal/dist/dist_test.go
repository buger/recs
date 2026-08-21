package dist_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"crm/internal/dist"
)

// Verifies: SYS-REQ-260821-AFPN SW-REQ-260821-AC3S STK-REQ-260820-T8AZ
// MCDC SYS-REQ-260821-AFPN: ape_binary_produced=F, ape_build_rejected=T, ape_build_requested=T, ape_runs_on_host=F, cosmocc_available=F, host_os_supported=F, native_blob_embedded=F, rc_GE_0=T => TRUE
// MCDC SW-REQ-260821-AC3S: ape_binary_produced=F, ape_build_rejected=T, ape_build_requested=T, ape_runs_on_host=F, cosmocc_available=F, host_os_supported=F, native_blob_embedded=F, rc_LE_0=F, rc_LE_255=F => TRUE
func TestAPEBuildFilesExist(t *testing.T) {
	if dist.BuildAPE() != "scripts/build-ape.sh" {
		t.Fatalf("build path %s", dist.BuildAPE())
	}
	root := findRoot(t)
	for _, p := range []string{"scripts/build-ape.sh", "ape/wrapper.c", "ape/hello.c", "Makefile", "docs/distribution.md"} {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Fatal(p, err)
		}
	}
	cmd := exec.Command("sh", filepath.Join(root, "scripts/build-ape.sh"))
	cmd.Env = []string{"PATH=/usr/bin:/bin", "COSMOCC_HOME=/tmp/crm-missing-cosmocc", "HOME=/tmp"}
	cmd.Dir = root
	if err := cmd.Run(); err == nil {
		t.Fatal("expected cosmocc-missing reject")
	}
}

func findRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("go.mod not found")
	return ""
}
