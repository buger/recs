package defaults

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed fs
var FS embed.FS

// WriteWorkspace copies embedded defaults if the target files do not exist.
// Implements: SYS-REQ-260820-KJ34
func WriteWorkspace(root string) error {
	return fs.WalkDir(FS, "fs", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(path, "fs/")
		dst := filepath.Join(root, rel)
		if _, err := os.Stat(dst); err == nil {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		data, err := FS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	})
}

// WriteAgentFiles writes AGENTS.md and SKILL.md into the workspace.
func WriteAgentFiles(root string) error {
	for _, name := range []string{"AGENTS.md", "SKILL.md"} {
		data, err := FS.ReadFile("fs/" + name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(root, name), data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
