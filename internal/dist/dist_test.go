package dist_test

import (
	"os"
	"path/filepath"
	"testing"
)

// Verifies: SYS-REQ-260821-AFPN SW-REQ-260821-AC3S STK-REQ-260820-T8AZ
func TestAPEBuildFilesExist(t *testing.T) {
	root := findRoot(t)
	for _, p := range []string{"scripts/build-ape.sh", "ape/wrapper.c", "ape/hello.c", "Makefile", "docs/distribution.md"} {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			t.Fatal(p, err)
		}
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
