package defaults_test

import (
	"os"
	"path/filepath"
	"testing"

	"crm/internal/defaults"
)

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
