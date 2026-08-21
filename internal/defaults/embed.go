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
		if err != nil { //mcdc:ignore:defensive embedded defaults tree cannot fail WalkDir
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
		if err != nil { //mcdc:ignore:defensive embedded file from WalkDir is always readable
			return err
		}
		return os.WriteFile(dst, data, 0o644)
	})
}

