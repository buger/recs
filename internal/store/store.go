package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"crm/internal/defaults"
	"crm/internal/record"
)

var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("conflict")
	ErrExists   = errors.New("record already exists")
)

// Store reads and writes Markdown records under a workspace root.
type Store struct {
	Root string
}

type Config struct {
	Name        string                       `yaml:"name"`
	DefaultPort int                          `yaml:"default_port"`
	Types       map[string]TypeSchema        `yaml:"types"`
}

type TypeSchema struct {
	Required []string               `yaml:"required"`
	Fields   map[string]FieldSchema `yaml:"fields"`
}

type FieldSchema struct {
	Type string   `yaml:"type"`
	Enum []string `yaml:"enum"`
}

// FindRoot walks up from start looking for crm.yaml.
// Implements: SYS-REQ-260820-9J7C
func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "crm.yaml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("crm.yaml not found from %s", start)
		}
		dir = parent
	}
}

// Init creates the Phase 1 workspace.
// Implements: SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2
func (s *Store) Init() error {
	if s.Root == "" {
		return fmt.Errorf("workspace root is empty")
	}
	dirs := []string{
		"records/people",
		"records/companies",
		"records/customers",
		"records/grants",
		"records/applications",
		"records/onboarding",
		"records/emails",
		"records/tasks",
		"records/misc",
		"boards",
		"inbox",
		"attachments",
		"templates",
		".crm/index",
		".crm/cache",
		".crm/runtime/locks",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(s.Root, d), 0o755); err != nil {
			return err
		}
	}
	if err := defaults.WriteWorkspace(s.Root); err != nil {
		return err
	}
	return nil
}

// Implements: SYS-REQ-260820-9J7C
func (s *Store) recordsDir() string { return filepath.Join(s.Root, "records") }

// LoadAll reads every Markdown record under records/.
// Implements: SYS-REQ-260820-9J7C
func (s *Store) LoadAll() ([]*record.Record, error) {
	var out []*record.Record
	root := s.recordsDir()
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		rec, err := s.readFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		out = append(out, rec)
		return nil
	})
	return out, err
}

// Implements: SYS-REQ-260820-9J7C
func (s *Store) readFile(path string) (*record.Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rec, err := record.Parse(path, data)
	if err != nil {
		return nil, err
	}
	if rec.ID == "" {
		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		rec.ID = base
		rec.Set("id", base)
	}
	return rec, nil
}

// Get returns the record with a stable id.
// Implements: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func (s *Store) Get(id string) (*record.Record, error) {
	recs, err := s.LoadAll()
	if err != nil {
		return nil, err
	}
	for _, rec := range recs {
		if rec.ID == id {
			return rec, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
}

// Create writes a new record file.
// Implements: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func (s *Store) Create(rec *record.Record) error {
	if rec.Type == "" {
		return fmt.Errorf("type is required")
	}
	if rec.ID == "" {
		title := rec.GetString("title")
		if title == "" {
			title = rec.GetString("name")
		}
		rec.ID = record.SlugID(rec.Type, title)
		rec.Set("id", rec.ID)
	}
	if rec.GetString("created_at") == "" {
		rec.Set("created_at", time.Now().UTC().Format(time.RFC3339))
	}
	rec.Set("updated_at", time.Now().UTC().Format(time.RFC3339))
	if rec.Path == "" {
		dir := filepath.Join(s.Root, "records", record.TypeDir(rec.Type))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		rec.Path = filepath.Join(dir, rec.ID+".md")
	}
	if _, err := os.Stat(rec.Path); err == nil {
		return fmt.Errorf("%w: %s", ErrExists, rec.ID)
	}
	if existing, err := s.Get(rec.ID); err == nil && existing != nil {
		return fmt.Errorf("%w: %s", ErrExists, rec.ID)
	} else if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	return s.Write(rec)
}

// Write stores a record with a temp file and rename.
// Implements: SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4 SYS-REQ-260820-7WT4 SW-REQ-260820-9C5Z
func (s *Store) Write(rec *record.Record) error {
	if rec.Path == "" {
		return fmt.Errorf("record path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(rec.Path), 0o755); err != nil {
		return err
	}
	data := rec.Bytes()
	tmp := rec.Path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, rec.Path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

type PatchResult struct {
	Record  *record.Record
	Changed map[string]map[string]any
	Version string
}

// Patch applies field updates with optional optimistic concurrency.
// Implements: SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4
func (s *Store) Patch(id string, sets map[string]any, deletes []string, ifVersion string) (*PatchResult, error) {
	lock, err := s.lock(id)
	if err != nil {
		return nil, err
	}
	defer lock.Close()

	rec, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	cur := rec.Version()
	if ifVersion != "" && ifVersion != cur {
		return nil, fmt.Errorf("%w: expected %s current %s", ErrConflict, ifVersion, cur)
	}
	changed := map[string]map[string]any{}
	for k, v := range sets {
		from := rec.Get(k)
		rec.Set(k, v)
		changed[k] = map[string]any{"from": from, "to": v}
	}
	for _, k := range deletes {
		from := rec.Get(k)
		rec.Delete(k)
		changed[k] = map[string]any{"from": from, "to": nil}
	}
	rec.Set("updated_at", time.Now().UTC().Format(time.RFC3339))
	if err := s.Write(rec); err != nil {
		return nil, err
	}
	return &PatchResult{Record: rec, Changed: changed, Version: rec.Version()}, nil
}

type fileLock struct {
	f *os.File
}

// Implements: SYS-REQ-260820-9J7C
func (l *fileLock) Close() {
	if l == nil || l.f == nil {
		return
	}
	_ = syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
	_ = l.f.Close()
}

// Implements: SYS-REQ-260820-9J7C
func (s *Store) lock(id string) (*fileLock, error) {
	dir := filepath.Join(s.Root, ".crm", "runtime", "locks")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	sum := sha256.Sum256([]byte(id))
	name := hex.EncodeToString(sum[:8]) + ".lock"
	f, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, err
	}
	return &fileLock{f: f}, nil
}

// Implements: SYS-REQ-260820-9J7C
func FileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
