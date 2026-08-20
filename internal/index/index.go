package index

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"crm/internal/record"
)

type Snapshot struct {
	Records   []map[string]any            `json:"records"`
	ByType    map[string][]string         `json:"by_type"`
	ByTag     map[string][]string         `json:"by_tag"`
	Backlinks map[string][]string         `json:"backlinks"`
}

// Rebuild writes disposable derived files.
// Implements: SYS-REQ-260820-Q8GR SW-REQ-260820-BNR7 INT-REQ-260820-2JKK
func Rebuild(root string, recs []*record.Record) (*Snapshot, error) {
	dir := filepath.Join(root, ".crm", "index")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	snap := &Snapshot{
		ByType:    map[string][]string{},
		ByTag:     map[string][]string{},
		Backlinks: map[string][]string{},
	}
	ids := map[string]bool{}
	for _, rec := range recs {
		ids[rec.ID] = true
	}
	for _, rec := range recs {
		item := map[string]any{"id": rec.ID, "type": rec.Type, "path": rec.Path, "status": rec.GetString("status")}
		if t := rec.GetString("title"); t != "" {
			item["title"] = t
		}
		if n := rec.GetString("name"); n != "" {
			item["name"] = n
		}
		snap.Records = append(snap.Records, item)
		snap.ByType[rec.Type] = append(snap.ByType[rec.Type], rec.ID)
		for _, tag := range record.StringSlice(rec.Get("tags")) {
			snap.ByTag[tag] = append(snap.ByTag[tag], rec.ID)
		}
		for _, target := range relatedIDs(rec) {
			if ids[target] && target != rec.ID {
				snap.Backlinks[target] = append(snap.Backlinks[target], rec.ID)
			}
		}
	}
	if err := writeJSON(filepath.Join(dir, "records.json"), snap.Records); err != nil {
		return nil, err
	}
	if err := writeJSON(filepath.Join(dir, "by-type.json"), snap.ByType); err != nil {
		return nil, err
	}
	if err := writeJSON(filepath.Join(dir, "by-tag.json"), snap.ByTag); err != nil {
		return nil, err
	}
	if err := writeJSON(filepath.Join(dir, "backlinks.json"), snap.Backlinks); err != nil {
		return nil, err
	}
	return snap, nil
}

// Implements: SYS-REQ-260820-Q8GR
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Implements: SYS-REQ-260820-Q8GR
func relatedIDs(rec *record.Record) []string {
	var out []string
	for _, key := range []string{"company", "customer", "owner", "technical_owner"} {
		if s := rec.GetString(key); s != "" && strings.Contains(s, "_") {
			out = append(out, s)
		}
	}
	for _, key := range []string{"people", "companies", "contacts", "related"} {
		out = append(out, record.StringSlice(rec.Get(key))...)
	}
	if rels, ok := rec.Get("relations").([]any); ok {
		for _, item := range rels {
			if m, ok := item.(map[string]any); ok {
				if t, ok := m["target"].(string); ok {
					out = append(out, t)
				}
			}
		}
	}
	return out
}
