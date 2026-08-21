package board

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/buger/recs/internal/query"
	"github.com/buger/recs/internal/record"
	"github.com/buger/recs/internal/store"
	"gopkg.in/yaml.v3"
)

type Board struct {
	File        string         `yaml:"-"`
	ID          string         `yaml:"id"`
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Match       any            `yaml:"match"`
	Columns     []Column       `yaml:"columns"`
	Column      ColumnConfig   `yaml:"column"`
	Filters        map[string]any  `yaml:"filters"`
	FilterControls []FilterControl `yaml:"filter_controls"`
}

type FilterControl struct {
	Field string `yaml:"field"`
	Type  string `yaml:"type"`
}

type ColumnConfig struct {
	Field string `yaml:"field"`
}

type Column struct {
	ID     string `yaml:"id"`
	Title  string `yaml:"title"`
	Match  any    `yaml:"match"`
	OnDrop *OnDrop `yaml:"on_drop"`
}

type OnDrop struct {
	Set    map[string]any `yaml:"set"`
	Remove []string       `yaml:"remove"`
}

type View struct {
	Board   *Board
	Columns []ColumnView
}

type ColumnView struct {
	Column  Column
	Records []*record.Record
}

// LoadAll reads boards/*.yaml.
// Implements: SYS-REQ-260820-4628 SW-REQ-260820-NBGR
func LoadAll(root string) ([]*Board, error) {
	dir := filepath.Join(root, "boards")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []*Board
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		b, err := Load(root, strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml"))
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, nil
}

// Implements: SYS-REQ-260820-4628
func Load(root, name string) (*Board, error) {
	if name == "" || strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return nil, fmt.Errorf("invalid board name %q", name)
	}
	for _, ext := range []string{".yaml", ".yml"} {
		path := filepath.Join(root, "boards", name+ext)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		var b Board
		if err := yaml.Unmarshal(data, &b); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		b.File = path
		if b.ID == "" {
			b.ID = name
		}
		if b.Column.Field == "" {
			b.Column.Field = "status"
		}
		return &b, nil
	}
	return nil, &store.ChoiceError{
		Code: "unknown_board", Field: "board", Value: name,
		Allowed: boardIDs(root),
		Msg:     fmt.Sprintf("board %s not found", name),
	}
}

// Implements: SYS-REQ-260820-4628
func (b *Board) Matches(rec *record.Record) bool {
	if b.Match == nil {
		return true
	}
	if !matchValue(rec, b.Match) {
		return false
	}
	for k, v := range b.Filters {
		if !fieldEquals(rec, k, v) {
			return false
		}
	}
	return true
}

// Implements: SYS-REQ-260820-4628
func (b *Board) ColumnOf(rec *record.Record) string {
	for _, col := range b.Columns {
		if col.Match != nil && matchValue(rec, col.Match) {
			return col.ID
		}
	}
	return rec.GetString(b.Column.Field)
}

// Implements: SYS-REQ-260820-4628
func (b *Board) ColumnByID(id string) (Column, bool) {
	for _, c := range b.Columns {
		if c.ID == id {
			return c, true
		}
	}
	return Column{}, false
}

// Project groups matching records into columns.
// Implements: SYS-REQ-260820-4628
func (b *Board) Project(recs []*record.Record) *View {
	v := &View{Board: b}
	by := map[string][]*record.Record{}
	for _, rec := range recs {
		if !b.Matches(rec) {
			continue
		}
		col := b.ColumnOf(rec)
		by[col] = append(by[col], rec)
	}
	seen := map[string]bool{}
	for _, col := range b.Columns {
		v.Columns = append(v.Columns, ColumnView{Column: col, Records: by[col.ID]})
		seen[col.ID] = true
	}
	for col, rs := range by {
		if !seen[col] && col != "" {
			v.Columns = append(v.Columns, ColumnView{Column: Column{ID: col, Title: col}, Records: rs})
		}
	}
	return v
}

// ApplyMove mutates frontmatter for a column drop.
// Implements: SYS-REQ-260820-BVBE SW-REQ-260820-EX7Q INT-REQ-260820-JRWN
func (b *Board) ApplyMove(rec *record.Record, columnID string) error {
	col, ok := b.ColumnByID(columnID)
	if !ok {
		return &store.ChoiceError{
			Code: "unknown_column", Field: "column", Value: columnID,
			Allowed: columnIDs(b),
			Msg:     fmt.Sprintf("unknown column %s", columnID),
		}
	}
	if col.OnDrop != nil {
		for k, v := range col.OnDrop.Set {
			rec.Set(k, expandValue(v))
		}
		for _, k := range col.OnDrop.Remove {
			rec.Delete(k)
		}
	} else {
		rec.Set(b.Column.Field, col.ID)
	}
	return nil
}

// Implements: SYS-REQ-260820-4628
func expandValue(v any) any {
	s, ok := v.(string)
	if !ok {
		return v
	}
	if s == "$now" {
		return time.Now().UTC().Format(time.RFC3339)
	}
	return s
}

// Implements: SYS-REQ-260820-4628
func matchValue(rec *record.Record, spec any) bool {
	switch t := spec.(type) {
	case map[string]any:
		if all, ok := t["all"]; ok {
			return matchAll(rec, all)
		}
		if anyv, ok := t["any"]; ok {
			return matchAny(rec, anyv)
		}
		for k, v := range t {
			if k == "all" || k == "any" { //mcdc:ignore:defensive all/any keys return before this loop
				continue
			}
			if !matchField(rec, k, v) {
				return false
			}
		}
		return true
	case []any:
		return matchAll(rec, t)
	default:
		return false
	}
}

// Implements: SYS-REQ-260820-4628
func matchAll(rec *record.Record, spec any) bool {
	items, ok := spec.([]any)
	if !ok {
		return matchValue(rec, spec)
	}
	for _, item := range items {
		if !matchValue(rec, item) {
			return false
		}
	}
	return true
}

// Implements: SYS-REQ-260820-4628
func matchAny(rec *record.Record, spec any) bool {
	items, ok := spec.([]any)
	if !ok {
		return matchValue(rec, spec)
	}
	for _, item := range items {
		if matchValue(rec, item) {
			return true
		}
	}
	return false
}

// Implements: SYS-REQ-260820-4628
func matchField(rec *record.Record, field string, spec any) bool {
	if m, ok := spec.(map[string]any); ok {
		if v, ok := m["in"]; ok {
			return inList(rec.Get(field), v)
		}
		if v, ok := m["not_in"]; ok {
			return !inList(rec.Get(field), v)
		}
		if v, ok := m["contains"]; ok {
			return query.Match(rec, []query.Clause{{Field: field, Op: "contains", Value: fmt.Sprint(v)}})
		}
		if v, ok := m["eq"]; ok {
			return fieldEquals(rec, field, v)
		}
		if _, ok := m["exists"]; ok {
			return rec.Get(field) != nil
		}
		if _, ok := m["missing"]; ok {
			return rec.Get(field) == nil || rec.GetString(field) == ""
		}
	}
	if sl, ok := spec.([]any); ok {
		return inList(rec.Get(field), sl)
	}
	return fieldEquals(rec, field, spec)
}

// Implements: SYS-REQ-260820-4628
func fieldEquals(rec *record.Record, field string, spec any) bool {
	got := rec.Get(field)
	if sl := record.StringSlice(spec); len(sl) > 1 {
		return inList(got, spec)
	}
	return strings.EqualFold(strings.TrimSpace(fmt.Sprint(got)), strings.TrimSpace(fmt.Sprint(spec)))
}

// Implements: SYS-REQ-260820-4628
func inList(got any, spec any) bool {
	var wants []string
	switch t := spec.(type) {
	case []any:
		for _, x := range t {
			wants = append(wants, fmt.Sprint(x))
		}
	case []string:
		wants = t
	default:
		wants = []string{fmt.Sprint(t)}
	}
	gots := record.StringSlice(got)
	if len(gots) == 0 {
		gots = []string{fmt.Sprint(got)}
	}
	for _, g := range gots {
		for _, w := range wants {
			if strings.EqualFold(g, w) {
				return true
			}
		}
	}
	return false
}

// Implements: SW-REQ-260821-E5V8
func boardIDs(root string) []string {
	entries, err := os.ReadDir(filepath.Join(root, "boards"))
	if err != nil {
		return []string{}
	}
	out := make([]string, 0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".yaml") {
			out = append(out, strings.TrimSuffix(name, ".yaml"))
			continue
		}
		if strings.HasSuffix(name, ".yml") {
			out = append(out, strings.TrimSuffix(name, ".yml"))
		}
	}
	return out
}

// Implements: SW-REQ-260821-E5V8
func columnIDs(b *Board) []string {
	out := make([]string, 0)
	for _, c := range b.Columns {
		out = append(out, c.ID)
	}
	return out
}
