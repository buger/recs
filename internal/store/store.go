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

	"gopkg.in/yaml.v3"

	"crm/internal/defaults"
	"crm/internal/record"
)

var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("conflict")
	ErrExists   = errors.New("record already exists")
)

// EnumError is a structured invalid_enum failure.
type EnumError struct {
	Field   string
	Value   string
	Allowed []string
}

func (e *EnumError) Error() string { return "invalid_enum" }

// ConflictError is a structured optimistic-concurrency failure.
type ConflictError struct {
	Expected string
	Current  string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s: expected %s current %s", ErrConflict, e.Expected, e.Current)
}

func (e *ConflictError) Unwrap() error { return ErrConflict }

// Store reads and writes Markdown records under a workspace root.
type Store struct {
	Root string
}

type Config struct {
	Name        string                `yaml:"name"`
	DefaultPort int                   `yaml:"default_port"`
	Types       map[string]TypeSchema `yaml:"types"`
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
	if err != nil { //mcdc:ignore:defensive filepath.Abs only fails when Getwd fails
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
		"dashboards",
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
	err := walkMarkdown(root, func(path string) error {
		rec, err := s.readFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		out = append(out, rec)
		return nil
	})
	if err != nil {
		return nil, err
	}
	inbox := filepath.Join(s.Root, "inbox")
	seen := map[string]bool{}
	for _, rec := range out {
		seen[rec.Path] = true
	}
	err = walkMarkdown(inbox, func(path string) error {
		if seen[path] { //mcdc:ignore:defensive inbox and records walks never share a path
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

func walkMarkdown(root string, fn func(path string) error) error {
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		return fn(path)
	})
}

// ApplyTemplate fills empty fields and body from templates/<type>.md.
// Implements: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func (s *Store) ApplyTemplate(rec *record.Record) {
	if rec == nil || rec.Type == "" {
		return
	}
	path := filepath.Join(s.Root, "templates", rec.Type+".md")
	if err := confinedToRoot(s.Root, path); err != nil {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	tpl, err := record.Parse(path, data)
	if err != nil {
		return
	}
	for k, v := range tpl.Fields {
		if k == "id" {
			continue
		}
		if rec.Get(k) == nil || rec.GetString(k) == "" {
			rec.Set(k, v)
		}
	}
	if strings.TrimSpace(rec.Body) == "" {
		body := tpl.Body
		title := rec.GetString("title")
		if title == "" {
			title = rec.GetString("name")
		}
		body = strings.ReplaceAll(body, "{{title}}", title)
		body = strings.ReplaceAll(body, "{{name}}", rec.GetString("name"))
		rec.Body = body
	}
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
	if !record.ValidType(rec.Type) {
		return fmt.Errorf("invalid type %q", rec.Type)
	}
	if rec.ID == "" {
		title := rec.GetString("title")
		if title == "" {
			title = rec.GetString("name")
		}
		rec.ID = record.SlugID(rec.Type, title)
		rec.Set("id", rec.ID)
	}
	if !record.ValidStableID(rec.ID) {
		return fmt.Errorf("invalid id %q", rec.ID)
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
	if existing, err := s.Get(rec.ID); err == nil && existing != nil { //mcdc:ignore:defensive Get never returns a nil record with a nil error
		return fmt.Errorf("%w: %s", ErrExists, rec.ID)
	} else if err != nil && !errors.Is(err, ErrNotFound) { //mcdc:ignore:defensive else-if is only reached when Get already returned an error
		return err
	}
	return s.Write(rec)
}

// Write stores a record with a temp file and rename.
// Implements: SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4 SYS-REQ-260820-7WT4 SW-REQ-260820-9C5Z
func (s *Store) Write(rec *record.Record) error {
	if rec.ID != "" {
		lock, err := s.lock(rec.ID)
		if err != nil {
			return err
		}
		defer lock.Close()
	}
	return s.writeLocked(rec)
}

// Implements: SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4 SYS-REQ-260820-9J7C
func (s *Store) writeLocked(rec *record.Record) error {
	if rec.Path == "" {
		return fmt.Errorf("record path is empty")
	}
	if err := confinedToRoot(s.Root, rec.Path); err != nil {
		return err
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

// Implements: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func confinedToRoot(root, path string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil { //mcdc:ignore:defensive filepath.Abs only fails when Getwd fails
		return err
	}
	absPath, err := filepath.Abs(path)
	if err != nil { //mcdc:ignore:defensive filepath.Abs only fails when Getwd fails
		return err
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil { //mcdc:ignore:defensive Rel of two absolute paths cannot fail
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("path escapes workspace: %s", path)
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
		return nil, &ConflictError{Expected: ifVersion, Current: cur}
	}
	changed := map[string]map[string]any{}
	if body, ok := sets["_body"]; ok {
		from := rec.Body
		rec.Body = fmt.Sprint(body)
		changed["_body"] = map[string]any{"from": from, "to": rec.Body}
		delete(sets, "_body")
	}
	for k, v := range sets {
		v = expandNow(v)
		from := rec.Get(k)
		rec.Set(k, v)
		changed[k] = map[string]any{"from": from, "to": v}
	}
	for _, k := range deletes {
		from := rec.Get(k)
		rec.Delete(k)
		changed[k] = map[string]any{"from": from, "to": nil}
	}
	if err := s.checkEnum(rec, sets); err != nil {
		return nil, err
	}
	rec.Set("updated_at", time.Now().UTC().Format(time.RFC3339))
	if err := s.writeLocked(rec); err != nil {
		return nil, err
	}
	return &PatchResult{Record: rec, Changed: changed, Version: rec.Version()}, nil
}

func (s *Store) checkEnum(rec *record.Record, sets map[string]any) error {
	cfg, err := loadTypeSchemas(s.Root)
	if err != nil || cfg == nil {
		return nil
	}
	schema, ok := cfg[rec.Type]
	if !ok {
		return nil
	}
	for k := range sets {
		fs, ok := schema[k]
		if !ok || len(fs) == 0 { //mcdc:ignore:defensive loadTypeSchemas only records non-empty enums
			continue
		}
		got := rec.GetString(k)
		if got == "" {
			continue
		}
		found := false
		for _, e := range fs {
			if strings.EqualFold(e, got) {
				found = true
				break
			}
		}
		if !found {
			return &EnumError{Field: k, Value: got, Allowed: fs}
		}
	}
	return nil
}

func loadTypeSchemas(root string) (map[string]map[string][]string, error) {
	data, err := os.ReadFile(filepath.Join(root, "crm.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var raw struct {
		Types map[string]struct {
			Fields map[string]struct {
				Enum []string `yaml:"enum"`
			} `yaml:"fields"`
		} `yaml:"types"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	out := map[string]map[string][]string{}
	for typ, ts := range raw.Types {
		fields := map[string][]string{}
		for name, fs := range ts.Fields {
			if len(fs.Enum) > 0 {
				fields[name] = fs.Enum
			}
		}
		if len(fields) > 0 {
			out[typ] = fields
		}
	}
	return out, nil
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
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil { //mcdc:ignore:defensive flock of a regular lock file created by this process does not fail
		_ = f.Close()
		return nil, err
	}
	return &fileLock{f: f}, nil
}


// Delete removes the record file after a lock.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (s *Store) Delete(id string) error {
	lock, err := s.lock(id)
	if err != nil {
		return err
	}
	defer lock.Close()
	rec, err := s.Get(id)
	if err != nil {
		return err
	}
	if rec.Path == "" {
		return fmt.Errorf("record path is empty")
	}
	if err := confinedToRoot(s.Root, rec.Path); err != nil {
		return err
	}
	if err := os.Remove(rec.Path); err != nil {
		return err
	}
	return nil
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

func expandNow(v any) any {
	s, ok := v.(string)
	if !ok {
		return v
	}
	if strings.EqualFold(s, "now") || s == "$now" {
		return time.Now().UTC().Format(time.RFC3339)
	}
	return v
}
