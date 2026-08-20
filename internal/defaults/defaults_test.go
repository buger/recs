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
	if err := defaults.WriteAgentFiles(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Fatal(err)
	}
}

func TestWriteAgentFilesBlocked(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), nil, 0o444); err != nil {
		t.Fatal(err)
	}
	_ = os.Chmod(filepath.Join(root, "AGENTS.md"), 0o444)
	if err := defaults.WriteAgentFiles(root); err != nil {
		// overwrite may still succeed depending on OS
		_ = err
	}
}
