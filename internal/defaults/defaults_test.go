package defaults_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buger/recs/internal/defaults"
)

// Verifies: SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2 SYS-REQ-260821-8FKR SW-REQ-260821-FCGM STK-REQ-260820-V5ZD
func TestWriteWorkspaceAndAgents(t *testing.T) {
	root := t.TempDir()
	if err := defaults.WriteWorkspace(root); err != nil {
		t.Fatal(err)
	}
	if err := defaults.WriteWorkspace(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "crm.yaml")); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"AGENTS.md", "SKILL.md"} {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			t.Fatalf("init must not write %s", name)
		}
	}
}
