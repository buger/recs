package record

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Record is one Markdown file with YAML frontmatter.
type Record struct {
	ID     string
	Type   string
	Path   string
	Fields map[string]any
	Body   string
}

// Parse reads a Markdown record.
// Implements: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func Parse(path string, data []byte) (*Record, error) {
	text := string(data)
	fields := map[string]any{}
	body := text
	if strings.HasPrefix(text, "---\n") || strings.HasPrefix(text, "---\r\n") {
		rest := text[4:]
		if strings.HasPrefix(text, "---\r\n") {
			rest = text[5:]
		}
		end := strings.Index(rest, "\n---\n")
		endCR := strings.Index(rest, "\n---\r\n")
		sep := "\n---\n"
		if end == -1 || (endCR != -1 && endCR < end) {
			end = endCR
			sep = "\n---\r\n"
		}
		if end == -1 {
			return nil, fmt.Errorf("missing closing frontmatter delimiter")
		}
		fm := rest[:end]
		body = rest[end+len(sep):]
		if err := yaml.Unmarshal([]byte(fm), &fields); err != nil {
			return nil, fmt.Errorf("frontmatter: %w", err)
		}
		if fields == nil {
			fields = map[string]any{}
		}
	}
	rec := &Record{Path: path, Fields: fields, Body: body}
	rec.ID = rec.GetString("id")
	rec.Type = rec.GetString("type")
	return rec, nil
}

// Bytes serializes frontmatter plus Markdown body.
// Implements: SYS-REQ-260820-9J7C
func (r *Record) Bytes() []byte {
	if r.Fields == nil {
		r.Fields = map[string]any{}
	}
	if r.ID != "" {
		r.Fields["id"] = r.ID
	}
	if r.Type != "" {
		r.Fields["type"] = r.Type
	}
	var b strings.Builder
	b.WriteString("---\n")
	enc, err := yaml.Marshal(orderedFields(r.Fields))
	if err == nil { //mcdc:ignore:defensive yaml.Marshal of string-keyed maps cannot fail
		b.Write(enc)
	}
	b.WriteString("---\n")
	b.WriteString(r.Body)
	if r.Body != "" && !strings.HasSuffix(r.Body, "\n") {
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

// Implements: SYS-REQ-260820-9J7C
func orderedFields(in map[string]any) yaml.Node {
	pref := []string{"id", "type", "name", "title", "status", "tags", "owner", "priority", "created_at", "updated_at"}
	seen := map[string]bool{}
	keys := make([]string, 0, len(in))
	for _, k := range pref {
		if _, ok := in[k]; ok {
			keys = append(keys, k)
			seen[k] = true
		}
	}
	rest := make([]string, 0)
	for k := range in {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	keys = append(keys, rest...)
	node := yaml.Node{Kind: yaml.MappingNode}
	for _, k := range keys {
		var kn, vn yaml.Node
		_ = kn.Encode(k)
		_ = vn.Encode(in[k])
		node.Content = append(node.Content, &kn, &vn)
	}
	return node
}

// Implements: SYS-REQ-260820-9J7C
func (r *Record) GetString(key string) string {
	v := r.Get(key)
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case time.Time:
		return t.UTC().Format(time.RFC3339)
	case fmt.Stringer:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}

// Get returns a field, including dotted paths such as amount.requested.
// Implements: SYS-REQ-260820-9J7C
func (r *Record) Get(key string) any {
	if r.Fields == nil {
		return nil
	}
	if v, ok := r.Fields[key]; ok {
		return v
	}
	cur := any(r.Fields)
	for _, part := range strings.Split(key, ".") {
		m, ok := asMap(cur)
		if !ok {
			return nil
		}
		cur, ok = m[part]
		if !ok {
			return nil
		}
	}
	return cur
}

// Set writes a field, including dotted paths.
// Implements: SYS-REQ-260820-2SQZ
func (r *Record) Set(key string, value any) {
	if r.Fields == nil {
		r.Fields = map[string]any{}
	}
	if !strings.Contains(key, ".") {
		r.Fields[key] = value
		if key == "id" {
			if s, ok := value.(string); ok {
				r.ID = s
			}
		}
		if key == "type" {
			if s, ok := value.(string); ok {
				r.Type = s
			}
		}
		return
	}
	parts := strings.Split(key, ".")
	cur := r.Fields
	for _, part := range parts[:len(parts)-1] {
		next, ok := asMap(cur[part])
		if !ok {
			next = map[string]any{}
			cur[part] = next
		}
		cur = next
	}
	cur[parts[len(parts)-1]] = value
}

// Implements: SYS-REQ-260820-9J7C
func (r *Record) Delete(key string) {
	if r.Fields == nil {
		return
	}
	if !strings.Contains(key, ".") {
		delete(r.Fields, key)
		return
	}
	parts := strings.Split(key, ".")
	cur := r.Fields
	for _, part := range parts[:len(parts)-1] {
		next, ok := asMap(cur[part])
		if !ok {
			return
		}
		cur = next
	}
	delete(cur, parts[len(parts)-1])
}

// Implements: SYS-REQ-260820-9J7C
func asMap(v any) (map[string]any, bool) {
	switch t := v.(type) {
	case map[string]any:
		return t, true
	case map[any]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[fmt.Sprint(k)] = val
		}
		return out, true
	default:
		return nil, false
	}
}

// Version is a derived content hash.
// Implements: SYS-REQ-260820-2SQZ SW-REQ-260820-Q3C4
func (r *Record) Version() string {
	sum := sha256.Sum256(r.Bytes())
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Implements: SYS-REQ-260820-9J7C
func ValidStableID(id string) bool {
	if id == "" || strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
		return false
	}
	return sanitizeID(id) == id
}

// Implements: SYS-REQ-260820-9J7C
func ValidType(typ string) bool {
	if typ == "" || strings.ContainsAny(typ, `/\`) || strings.Contains(typ, "..") {
		return false
	}
	return sanitizeID(typ) == typ
}

// Implements: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func SlugID(typ, title string) string {
	typ = sanitizeID(typ)
	title = sanitizeID(title)
	if title == "" {
		title = fmt.Sprintf("%d", time.Now().UTC().UnixNano()%1_000_000)
	}
	if typ == "" {
		typ = "record"
	}
	return typ + "_" + title
}

// Implements: SYS-REQ-260820-9J7C
func sanitizeID(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('_')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "_")
}

// Implements: SYS-REQ-260820-9J7C
func TypeDir(typ string) string {
	switch typ {
	case "person":
		return "people"
	case "company":
		return "companies"
	case "customer":
		return "customers"
	case "grant":
		return "grants"
	case "application":
		return "applications"
	case "onboarding":
		return "onboarding"
	case "email":
		return "emails"
	case "task":
		return "tasks"
	case "note":
		return "notes"
	case "inbox":
		return "inbox"
	default:
		if typ == "" {
			return "misc"
		}
		return typ + "s"
	}
}

// Implements: SYS-REQ-260820-9J7C
func DisplayName(r *Record) string {
	for _, k := range []string{"title", "name", "subject"} {
		if s := r.GetString(k); s != "" {
			return s
		}
	}
	if r.ID != "" {
		return r.ID
	}
	return filepath.Base(r.Path)
}

// Implements: SYS-REQ-260820-9J7C
func CompareValues(a, b any) int {
	as, aNum, aOk := asFloat(a)
	bs, bNum, bOk := asFloat(b)
	if aOk && bOk {
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
		return 0
	}
	if at, ok := asTime(a); ok {
		if bt, ok2 := asTime(b); ok2 {
			if at.Before(bt) {
				return -1
			}
			if at.After(bt) {
				return 1
			}
			return 0
		}
	}
	return strings.Compare(strings.ToLower(fmt.Sprint(as)), strings.ToLower(fmt.Sprint(bs)))
}

// Implements: SYS-REQ-260820-9J7C
func asFloat(v any) (string, float64, bool) {
	s := strings.TrimSpace(fmt.Sprint(v))
	if v == nil {
		return "", 0, false
	}
	switch t := v.(type) {
	case int:
		return s, float64(t), true
	case int64:
		return s, float64(t), true
	case float64:
		return s, t, true
	case float32:
		return s, float64(t), true
	case uint64:
		return s, float64(t), true
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return s, 0, false
	}
	return s, f, true
}

// Implements: SYS-REQ-260820-9J7C
func asTime(v any) (time.Time, bool) {
	s := strings.TrimSpace(fmt.Sprint(v))
	for _, layout := range []string{time.RFC3339, "2006-01-02", "2006-01-02T15:04:05Z"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// Implements: SYS-REQ-260820-9J7C
func StringSlice(v any) []string {
	switch t := v.(type) {
	case nil:
		return nil
	case []string:
		return t
	case []any:
		out := make([]string, 0, len(t))
		for _, x := range t {
			out = append(out, fmt.Sprint(x))
		}
		return out
	default:
		s := strings.TrimSpace(fmt.Sprint(t))
		if s == "" {
			return nil
		}
		return []string{s}
	}
}
