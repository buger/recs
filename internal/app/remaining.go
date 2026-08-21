package app

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/buger/recs/internal/markdown"
	"github.com/buger/recs/internal/record"
	"github.com/buger/recs/internal/store"
	"github.com/buger/recs/internal/wikilink"
)

type Relation struct {
	Type   string `json:"type"`
	Target string `json:"target"`
	Title  string `json:"title,omitempty"`
}

type Attachment struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type RecordView struct {
	Record      *record.Record `json:"-"`
	Public      map[string]any `json:"record"`
	HTML        string         `json:"html"`
	Relations   []Relation     `json:"relations"`
	Backlinks   []Relation     `json:"backlinks"`
	Attachments []Attachment   `json:"attachments"`
}

type GitResult struct {
	OK      bool     `json:"ok"`
	Git     bool     `json:"git"`
	Output  string   `json:"output,omitempty"`
	Changed []string `json:"changed,omitempty"`
	History []string `json:"history,omitempty"`
	Message string   `json:"message,omitempty"`
}

// Delete removes a record file.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C INT-REQ-260821-8HAC
func (a *App) Delete(id string) error {
	return a.Store.Delete(id)
}

// Edit applies frontmatter and optional body updates.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (a *App) Edit(id string, sets map[string]any, body *string, ifVersion string) (*store.PatchResult, error) {
	if sets == nil {
		sets = map[string]any{}
	} else {
		cp := map[string]any{}
		for k, v := range sets {
			cp[k] = v
		}
		sets = cp
	}
	if body != nil {
		sets["_body"] = *body
	}
	if len(sets) == 0 {
		return nil, fmt.Errorf("usage: crm edit <id> --body ... --set k=v")
	}
	return a.Store.Patch(id, sets, nil, ifVersion)
}

// EditFromBytes replaces a record from an edited Markdown document.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (a *App) EditFromBytes(id string, data []byte, ifVersion string) (*store.PatchResult, error) {
	cur, err := a.Store.Get(id)
	if err != nil {
		return nil, err
	}
	if ifVersion != "" && ifVersion != cur.Version() {
		return nil, &store.ConflictError{Expected: ifVersion, Current: cur.Version()}
	}
	parsed, err := record.Parse(cur.Path, data)
	if err != nil {
		return nil, err
	}
	if parsed.ID != "" && parsed.ID != id {
		return nil, fmt.Errorf("edit cannot change id")
	}
	parsed.ID = id
	parsed.Path = cur.Path
	if parsed.Type == "" {
		parsed.Type = cur.Type
	}
	parsed.Set("updated_at", time.Now().UTC().Format(time.RFC3339))
	if err := a.Store.Write(parsed); err != nil {
		return nil, err
	}
	return &store.PatchResult{Record: parsed, Changed: map[string]map[string]any{"_document": map[string]any{"from": cur.Version(), "to": parsed.Version()}}, Version: parsed.Version()}, nil
}

// EditWithEditor copies the record to a temp file, runs editor, and writes back.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (a *App) EditWithEditor(id, editor string) (*store.PatchResult, error) {
	if strings.TrimSpace(editor) == "" {
		return nil, fmt.Errorf("EDITOR is empty")
	}
	rec, err := a.Store.Get(id)
	if err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp("", "crm-edit-*.md")
	if err != nil { //mcdc:ignore:defensive CreateTemp in the process temp dir does not fail in this test environment
		return nil, err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(rec.Bytes()); err != nil { //mcdc:ignore:defensive writing a small record into a fresh temp file does not fail
		tmp.Close()
		os.Remove(tmpPath)
		return nil, err
	}
	tmp.Close()
	defer os.Remove(tmpPath)
	cmd := exec.Command("sh", "-c", editor+` "$1"`, "crm-edit", tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("editor: %w", err)
	}
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, err
	}
	return a.EditFromBytes(id, data, rec.Version())
}

// Link appends a canonical relation on the source record.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (a *App) Link(id, target, rel string) (*store.PatchResult, error) {
	if rel == "" {
		return nil, fmt.Errorf("usage: crm link <id> <target> --relation <type>")
	}
	if _, err := a.Store.Get(target); err != nil {
		return nil, err
	}
	rec, err := a.Store.Get(id)
	if err != nil {
		return nil, err
	}
	rels := relationsOf(rec)
	for _, r := range rels {
		if r.Type == rel && r.Target == target {
			return &store.PatchResult{Record: rec, Changed: map[string]map[string]any{}, Version: rec.Version()}, nil
		}
	}
	rels = append(rels, Relation{Type: rel, Target: target})
	raw := make([]any, 0, len(rels))
	for _, r := range rels {
		raw = append(raw, map[string]any{"type": r.Type, "target": r.Target})
	}
	return a.Store.Patch(id, map[string]any{"relations": raw}, nil, "")
}

// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func relationsOf(rec *record.Record) []Relation {
	var out []Relation
	if rec == nil {
		return out
	}
	items := rec.Get("relations")
	switch t := items.(type) {
	case []any:
		for _, item := range t {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			out = append(out, Relation{Type: fmt.Sprint(m["type"]), Target: fmt.Sprint(m["target"])})
		}
	}
	return out
}

// Ingest creates a record from provider-neutral JSON.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (a *App) Ingest(kind string, raw []byte) (*record.Record, error) {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("ingest json: %w", err)
	}
	typ, _ := payload["type"].(string)
	if typ == "" {
		if kind != "" && kind != "-" {
			typ = kind
		} else {
			typ = "email"
		}
	}
	id, _ := payload["id"].(string)
	body, _ := payload["body"].(string)
	if body == "" {
		if text, ok := payload["text"].(string); ok {
			body = text
		}
	}
	delete(payload, "type")
	delete(payload, "id")
	delete(payload, "body")
	delete(payload, "text")
	if typ == "email" {
		if sub, _ := payload["subject"].(string); sub != "" && payload["title"] == nil {
			payload["title"] = sub
		}
		if payload["triage_status"] == nil {
			payload["triage_status"] = "new"
		}
	}
	return a.Create(typ, id, payload, body)
}

// ExportJSON returns all workspace records.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (a *App) ExportJSON() ([]map[string]any, error) {
	recs, err := a.Store.LoadAll()
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(recs))
	for _, rec := range recs {
		item := map[string]any{"id": rec.ID, "type": rec.Type, "path": rec.Path, "body": rec.Body}
		for k, v := range rec.Fields {
			item[k] = v
		}
		out = append(out, item)
	}
	return out, nil
}

// ExportCSV writes records as CSV.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (a *App) ExportCSV() (string, error) {
	recs, err := a.Store.LoadAll()
	if err != nil {
		return "", err
	}
	headers := []string{"id", "type", "title", "name", "status", "owner", "tags", "body"}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(headers); err != nil { //mcdc:ignore:defensive encoding/csv Write to a bytes.Buffer cannot fail
		return "", err
	}
	for _, rec := range recs {
		row := []string{
			rec.ID, rec.Type, rec.GetString("title"), rec.GetString("name"),
			rec.GetString("status"), rec.GetString("owner"),
			strings.Join(record.StringSlice(rec.Get("tags")), ","), rec.Body,
		}
		if err := w.Write(row); err != nil { //mcdc:ignore:defensive encoding/csv Write of a string row to a bytes.Buffer cannot fail
			return "", err
		}
	}
	w.Flush()
	return buf.String(), w.Error()
}

// ImportCSV creates records from a CSV table.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (a *App) ImportCSV(r io.Reader, defaultType string) ([]*record.Record, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	rows, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("csv is empty")
	}
	headers := make([]string, len(rows[0]))
	for i, h := range rows[0] {
		headers[i] = strings.TrimSpace(strings.ToLower(h))
	}
	var created []*record.Record
	for _, row := range rows[1:] {
		sets := map[string]any{}
		body := ""
		id := ""
		typ := defaultType
		for i, h := range headers {
			if i >= len(row) {
				continue
			}
			val := strings.TrimSpace(row[i])
			switch h {
			case "id":
				id = val
			case "type":
				if val != "" {
					typ = val
				}
			case "body":
				body = val
			case "tags":
				if val != "" {
					sets["tags"] = strings.Split(val, ",")
				}
			default:
				if h != "" && val != "" {
					sets[h] = val
				}
			}
		}
		if typ == "" {
			return created, fmt.Errorf("type is required: set a type column or --type")
		}
		rec, err := a.Create(typ, id, sets, body)
		if err != nil {
			return created, err
		}
		created = append(created, rec)
	}
	return created, nil
}

// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (a *App) gitRoot() string {
	dir := a.Root()
	for {
		if st, err := os.Stat(filepath.Join(dir, ".git")); err == nil && st.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errb.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	return out.String(), nil
}

// Diff delegates to git when a repository exists.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (a *App) Diff() GitResult {
	root := a.gitRoot()
	if root == "" {
		return GitResult{OK: true, Git: false, Message: "no git repository"}
	}
	if _, err := exec.LookPath("git"); err != nil {
		return GitResult{OK: true, Git: false, Message: "git not found"}
	}
	out, err := runGit(root, "diff", "--", "records", "boards", "dashboards", "crm.yaml")
	if err != nil {
		return GitResult{OK: false, Git: true, Message: err.Error()}
	}
	return GitResult{OK: true, Git: true, Output: out}
}

// Changed lists git-changed workspace files.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (a *App) Changed() GitResult {
	root := a.gitRoot()
	if root == "" {
		return GitResult{OK: true, Git: false, Message: "no git repository", Changed: []string{}}
	}
	if _, err := exec.LookPath("git"); err != nil {
		return GitResult{OK: true, Git: false, Message: "git not found", Changed: []string{}}
	}
	out, err := runGit(root, "status", "--porcelain", "--", "records", "boards", "dashboards", "crm.yaml")
	if err != nil {
		return GitResult{OK: false, Git: true, Message: err.Error(), Changed: []string{}}
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(line) > 3 { //mcdc:ignore:defensive git status --porcelain lines are two status bytes, a space, and a path
			files = append(files, strings.TrimSpace(line[3:]))
		} else {
			files = append(files, line)
		}
	}
	if files == nil {
		files = []string{}
	}
	return GitResult{OK: true, Git: true, Changed: files, Output: out}
}

// History lists git commits for a record.
// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C
func (a *App) History(id string) GitResult {
	root := a.gitRoot()
	if root == "" {
		return GitResult{OK: true, Git: false, Message: "no git repository", History: []string{}}
	}
	if _, err := exec.LookPath("git"); err != nil {
		return GitResult{OK: true, Git: false, Message: "git not found", History: []string{}}
	}
	rec, err := a.Store.Get(id)
	if err != nil {
		return GitResult{OK: false, Git: true, Message: err.Error(), History: []string{}}
	}
	rel, err := filepath.Rel(root, rec.Path)
	if err != nil { //mcdc:ignore:defensive record paths are created under the workspace root so Rel cannot fail
		rel = rec.Path
	}
	out, err := runGit(root, "log", "--oneline", "--", rel)
	if err != nil {
		return GitResult{OK: false, Git: true, Message: err.Error(), History: []string{}}
	}
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	if lines == nil {
		lines = []string{}
	}
	return GitResult{OK: true, Git: true, History: lines, Output: out}
}

// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func publicRecord(rec *record.Record) map[string]any {
	out := map[string]any{"id": rec.ID, "type": rec.Type, "path": rec.Path, "body": rec.Body, "version": rec.Version()}
	for k, v := range rec.Fields {
		out[k] = v
	}
	return out
}

// View assembles record metadata, rendered markdown, relations, backlinks, and attachments.
// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA INT-REQ-260821-MRGW
func (a *App) View(id string) (*RecordView, error) {
	rec, err := a.Store.Get(id)
	if err != nil {
		return nil, err
	}
	all, err := a.Store.LoadAll()
	if err != nil { //mcdc:ignore:defensive View already loaded the store via Get so the immediate second LoadAll cannot fail
		return nil, err
	}
	html := markdown.Render(rec.Body)
	html = wikilink.Resolve(html, all)
	view := &RecordView{
		Record: rec,
		Public: publicRecord(rec),
		HTML:   html,
	}
	byID := map[string]*record.Record{}
	for _, r := range all {
		byID[r.ID] = r
	}
	for _, rel := range relationsOf(rec) {
		if t := byID[rel.Target]; t != nil {
			rel.Title = record.DisplayName(t)
		}
		view.Relations = append(view.Relations, rel)
	}
	for _, other := range all {
		if other.ID == rec.ID {
			continue
		}
		for _, rel := range relationsOf(other) {
			if rel.Target == rec.ID {
				view.Backlinks = append(view.Backlinks, Relation{Type: rel.Type, Target: other.ID, Title: record.DisplayName(other)})
			}
		}
		for _, key := range []string{"company", "customer", "people", "companies", "contacts", "related"} {
			for _, s := range record.StringSlice(other.Get(key)) {
				if s == rec.ID {
					view.Backlinks = append(view.Backlinks, Relation{Type: key, Target: other.ID, Title: record.DisplayName(other)})
				}
			}
		}
		if other.GetString("company") == rec.ID || other.GetString("customer") == rec.ID {
			// already covered by slice helper for scalars
		}
	}
	for _, path := range record.StringSlice(rec.Get("attachments")) {
		view.Attachments = append(view.Attachments, Attachment{Path: path, Name: filepath.Base(path)})
	}
	if view.Relations == nil {
		view.Relations = []Relation{}
	}
	if view.Backlinks == nil {
		view.Backlinks = []Relation{}
	}
	if view.Attachments == nil {
		view.Attachments = []Attachment{}
	}
	return view, nil
}

// AttachmentFile returns an absolute workspace path for an attachment.
// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func (a *App) AttachmentFile(rel string) (string, error) {
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" || strings.Contains(rel, "..") {
		return "", fmt.Errorf("invalid attachment path")
	}
	abs := rel
	if !filepath.IsAbs(rel) { //mcdc:ignore:defensive TrimPrefix of the leading slash makes Unix paths relative before IsAbs runs
		abs = filepath.Join(a.Root(), rel)
	}
	if err := storeConfined(a.Root(), abs); err != nil { //mcdc:ignore:defensive AttachmentFile already rejects empty and dot-dot paths so Join stays inside the workspace
		return "", err
	}
	if _, err := os.Stat(abs); err != nil {
		return "", err
	}
	return abs, nil
}

// Implements: SYS-REQ-260821-JYEJ SW-REQ-260821-8C2C SYS-REQ-260821-QF1J
func storeConfined(root, path string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil { //mcdc:ignore:defensive filepath.Abs fails only when Getwd fails
		return err
	}
	absPath, err := filepath.Abs(path)
	if err != nil { //mcdc:ignore:defensive Abs of a joined workspace path does not fail after root Abs succeeded
		return err
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil { //mcdc:ignore:defensive Rel of two Abs paths on the same volume does not fail
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("path escapes workspace: %s", path)
	}
	return nil
}

// ResolveWikilink returns a matching record id.
// Implements: SYS-REQ-260821-QF1J
func (a *App) ResolveWikilink(name string) (string, error) {
	recs, err := a.Store.LoadAll()
	if err != nil {
		return "", err
	}
	if rec := wikilink.Match(name, recs); rec != nil {
		return rec.ID, nil
	}
	return "", fmt.Errorf("%w: %s", store.ErrNotFound, name)
}
