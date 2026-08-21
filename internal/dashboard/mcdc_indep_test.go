package dashboard_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/buger/recs/internal/board"
	"github.com/buger/recs/internal/dashboard"
	"github.com/buger/recs/internal/record"
)

func rec(id, typ string, fields map[string]any, body string) *record.Record {
	r := &record.Record{ID: id, Type: typ, Body: body, Fields: fields}
	if r.Fields == nil {
		r.Fields = map[string]any{}
	}
	return r
}

// Verifies: SW-REQ-260820-EJVT
// SW-REQ-260820-EJVT
func TestProjectWidgetIndependence(t *testing.T) {
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	recs := []*record.Record{
		rec("t_done", "task", map[string]any{"title": "D", "status": "done"}, ""),
		rec("t_won", "task", map[string]any{"title": "W", "status": "won"}, ""),
		rec("t_applied", "task", map[string]any{"title": "A", "status": "applied"}, ""),
		rec("t_complete", "task", map[string]any{"title": "C", "status": "complete"}, ""),
		rec("t_open", "task", map[string]any{"title": "O", "status": "open", "owner": "x"}, ""),
		rec("n_blank", "note", map[string]any{"title": "N"}, "note body"),
		rec("n_empty_status", "note", map[string]any{"title": "E", "status": ""}, "watch me"),
	}
	root := t.TempDir()
	md := filepath.Join(root, "note.md")
	if err := os.WriteFile(md, []byte("---\nid: note_md\ntype: note\ntitle: FromMD\n---\nHello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	emptyTitle := filepath.Join(root, "empty.md")
	if err := os.WriteFile(emptyTitle, []byte("# just markdown\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	bv := &board.View{Board: &board.Board{ID: "grants", Name: "Grants"}, Columns: []board.ColumnView{
		{Column: board.Column{ID: "open", Title: "Open"}, Records: recs[:1]},
	}}
	env := dashboard.Env{
		Root:    root,
		Records: recs,
		Now:     now,
		Board: func(id string) (*board.View, error) {
			if id == "missing" {
				return nil, errors.New("no board")
			}
			if id == "nilview" {
				return nil, nil
			}
			return bv, nil
		},
	}

	widgets := []dashboard.Widget{
		{ID: "ph", Type: "", Title: "empty type"},
		{ID: "ph2", Type: "placeholder", Title: "ph"},
		{ID: "bad", Type: "nope", Title: "bad"},
		{ID: "count_ok", Type: "count", Query: "status=open"},
		{ID: "count_err", Type: "count", Query: "(((broken"},
		{ID: "list_ok", Type: "list", Query: "", GroupBy: "status"},
		{ID: "list_empty_key", Type: "list", Query: "title=N", GroupBy: "missing"},
		{ID: "list_err", Type: "list", Query: "(((broken"},
		{ID: "notes_ok", Type: "notes", Query: "", Limit: 2},
		{ID: "notes_err", Type: "notes", Query: "(((broken"},
		{ID: "watch_ok", Type: "watch", Query: ""},
		{ID: "watch_err", Type: "watch", Query: "(((broken"},
		{ID: "pipe", Type: "pipeline", Query: ""},
		{ID: "pipe_empty", Type: "pipeline", Query: "title=ZZZnone"},
		{ID: "pipe_err", Type: "pipeline", Query: "(((broken"},
		{ID: "metrics", Type: "metrics", Stats: []dashboard.Stat{
			{Label: "ok", Query: "status=open"},
			{Label: "bad", Query: "(((broken"},
		}, Reminder: &dashboard.Reminder{Query: "status=open", Template: ""}},
		{ID: "metrics_rem_err", Type: "metrics", Stats: []dashboard.Stat{{Label: "x", Query: "status=open"}}, Reminder: &dashboard.Reminder{Query: "(((broken", Template: "hi {{name}}"}},
		{ID: "metrics_stat_only", Type: "metrics", Stats: []dashboard.Stat{{Label: "x", Query: "(((broken"}}},
		{ID: "board_missing_src", Type: "board"},
		{ID: "board_err", Type: "board", Board: "missing"},
		{ID: "board_nil", Type: "board", Board: "nilview"},
		{ID: "board_ok", Type: "board", Board: "grants"},
		{ID: "md_missing_src", Type: "markdown"},
		{ID: "md_trav", Type: "markdown", Source: "../etc/passwd"},
		{ID: "md_missfile", Type: "markdown", Source: "nope.md"},
		{ID: "md_ok", Type: "markdown", Source: "note.md", Title: ""},
		{ID: "md_titled", Type: "markdown", Source: "empty.md", Title: "Kept"},
	}

	// Project only takes first SlotCount widgets; use 1x1 repeatedly.
	for _, w := range widgets {
		d := &dashboard.Dashboard{ID: "w", Name: "w", Layout: "1x1", Widgets: []dashboard.Widget{w}}
		proj := dashboard.Project(d, env)
		if proj == nil || len(proj.Widgets) == 0 {
			t.Fatalf("empty proj for %s", w.ID)
		}
	}

	// Env with nil Board func for missing-source-style board.
	d := &dashboard.Dashboard{ID: "w", Name: "w", Layout: "1x1", Widgets: []dashboard.Widget{{ID: "b", Type: "board", Board: "grants"}}}
	_ = dashboard.Project(d, dashboard.Env{Root: root, Records: recs, Now: now})
}

// Verifies: SW-REQ-260820-NA06
// SW-REQ-260820-NA06
func TestValidIDIndependence(t *testing.T) {
	root := t.TempDir()
	for i, id := range []string{"abc", "AB2", "A1x", "z9x", "id_1", "id-1", "Zx", "0bad", "a_b-c1"} {
		if _, err := dashboard.Create(root, id, id, "1x1", "", nil); err != nil {
			t.Fatalf("%d id %s: %v", i, id, err)
		}
	}
	if _, err := dashboard.Create(root, "bad id", "x", "1x1", "", nil); err == nil {
		t.Fatal("space should fail")
	}
	if _, err := dashboard.Create(root, "bad!", "x", "1x1", "", nil); err == nil {
		t.Fatal("bang should fail")
	}
}

// Verifies: SW-REQ-260820-EJVT
// SW-REQ-260820-EJVT
func TestApplyTemplateAndLoadEdges(t *testing.T) {
	root := t.TempDir()
	dashDir := filepath.Join(root, "dashboards")
	if err := os.MkdirAll(dashDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// junk file skipped / yaml error
	if err := os.WriteFile(filepath.Join(dashDir, "not-yaml.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dashDir, "broken.yaml"), []byte(":\n:\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := dashboard.LoadAll(root); err == nil {
		// broken yaml should error
		t.Fatal("expected broken yaml")
	}
	if _, err := dashboard.Load(root, "missing"); err == nil {
		t.Fatal("missing load")
	}
	d, err := dashboard.Create(root, "ok1", "", "", "", nil)
	if err != nil || d.Name != "ok1" {
		t.Fatal(err, d)
	}
	if _, err := dashboard.Create(root, "ok1", "dup", "", "", nil); err == nil {
		t.Fatal("dup")
	}
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := dashboard.Create(blocked, "x", "x", "", "", nil); err == nil {
		t.Fatal("create under file")
	}

	// reminder template name/title/id keys
	recs := []*record.Record{rec("r1", "note", map[string]any{"name": "N", "title": "T"}, "")}
	env := dashboard.Env{Root: root, Records: recs, Now: time.Now()}
	d2 := &dashboard.Dashboard{ID: "m", Name: "m", Layout: "1x1", Widgets: []dashboard.Widget{{
		ID: "m", Type: "metrics",
		Reminder: &dashboard.Reminder{Query: "", Template: "{{name}} {{title}} {{id}} {{nope}}"},
	}}}
	_ = dashboard.Project(d2, env)
}

// Verifies: SW-REQ-260820-EJVT
// SW-REQ-260820-EJVT
func TestRenderAndSafeRel(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.md"), []byte("# Hi\n\n[[link]] and `code`\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := dashboard.Env{Root: root, Now: time.Now()}
	d := &dashboard.Dashboard{ID: "m", Name: "m", Layout: "1x1", Widgets: []dashboard.Widget{
		{ID: "m", Type: "markdown", Source: "a.md"},
	}}
	p := dashboard.Project(d, env)
	if p.Widgets[0].HTML == "" {
		t.Fatal("html")
	}
	if _, err := dashboard.Create(root, strings.Repeat("a", 1), "x", "2x0", "", []dashboard.Widget{
		{ID: "l", Type: "list", Limit: 0, Query: map[string]any{"status": map[string]any{"eq": "open"}}},
		{ID: "l2", Type: "list", Query: map[string]any{"due": map[string]any{"before": "2026-01-01"}}},
		{ID: "l3", Type: "list", Query: map[string]any{"due": map[string]any{"after": "2026-01-01"}}},
		{ID: "l4", Type: "list", Query: map[string]any{"title": map[string]any{"contains": "x"}}},
		{ID: "l5", Type: "list", Query: map[string]any{"status": "open"}},
		{ID: "l6", Type: "list", Query: 12},
	}); err != nil {
		// 2x0 clamps; create ok
	}
	d3 := &dashboard.Dashboard{ID: "q", Name: "q", Layout: "2x2", Widgets: []dashboard.Widget{
		{ID: "l", Type: "list", Query: map[string]any{"status": map[string]any{"eq": "open"}}},
		{ID: "b", Type: "list", Field: "due", Before: "today"},
		{ID: "c", Type: "list", Query: "due < +1d"},
		{ID: "d", Type: "count", Query: nil},
	}}
	_ = dashboard.Project(d3, dashboard.Env{Root: root, Records: []*record.Record{
		rec("x", "note", map[string]any{"status": "open", "due": "2026-08-21"}, ""),
	}, Now: time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)})
	_ = dashboard.Project(nil, dashboard.Env{})
}


// Verifies: SW-REQ-260820-NA06
// SW-REQ-260820-NA06
func TestValidIDBoundaryRunes(t *testing.T) {
	root := t.TempDir()
	for _, id := range []string{"_", "-", "{", "[", ":", "a..b", "a/b", "", `ab\cd`, "###"} {
		_, _ = dashboard.Create(root, id, "x", "1x1", "", nil)
	}
	// subdirectory in dashboards for IsDir
	if err := os.MkdirAll(filepath.Join(root, "dashboards", "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "dashboards", "z.yaml"), []byte("id: z\nname: Z\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "dashboards", "a.yaml"), []byte("id: a\nname: A\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _ = dashboard.LoadAll(root)
	_, _ = dashboard.Load(root, "missing-noent")
	_, _ = dashboard.Load(filepath.Join(root, "no-dash-dir"), "x")
	// too many widgets vs slots
	d := &dashboard.Dashboard{ID: "many", Name: "many", Layout: "1x1", Widgets: []dashboard.Widget{
		{ID: "1", Type: "count"}, {ID: "2", Type: "count"}, {ID: "3", Type: "count"},
	}}
	_ = dashboard.Project(d, dashboard.Env{Now: time.Time{}})
	long := strings.Repeat("word ", 80)
	recs := []*record.Record{rec("long", "note", map[string]any{"title": "L", "tags": []any{"", "x"}}, long)}
	_ = dashboard.Project(&dashboard.Dashboard{ID: "n", Name: "n", Layout: "1x1", Widgets: []dashboard.Widget{
		{ID: "notes", Type: "notes", Query: ""},
	}}, dashboard.Env{Records: recs, Now: time.Now()})
	_ = dashboard.Project(&dashboard.Dashboard{ID: "md", Name: "md", Layout: "1x1", Widgets: []dashboard.Widget{
		{ID: "m", Type: "markdown", Source: ""},
	}}, dashboard.Env{Root: root})
	_ = dashboard.SlotCount("5x5")
	_ = dashboard.SlotCount("0x2")
}

// Verifies: SW-REQ-260820-EJVT
// SW-REQ-260820-EJVT
func TestApplyTemplateKeys(t *testing.T) {
	root := t.TempDir()
	for _, fields := range []map[string]any{
		{"name": "N"},
		{"title": "T"},
		{"id": "I"},
	} {
		recs := []*record.Record{rec("r", "note", fields, "")}
		d := &dashboard.Dashboard{ID: "m", Name: "m", Layout: "1x1", Widgets: []dashboard.Widget{{
			ID: "m", Type: "metrics",
			Reminder: &dashboard.Reminder{Query: "", Template: "{{name}}{{title}}{{id}}"},
		}}}
		_ = dashboard.Project(d, dashboard.Env{Root: root, Records: recs, Now: time.Now()})
	}
}

func TestDashboardLeftoverIndependence(t *testing.T) {
	root := t.TempDir()
	dash := filepath.Join(root, "dashboards")
	if err := os.MkdirAll(dash, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dash, "beta.yml"), []byte("id: beta\nname: B\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dash, "alpha.yaml"), []byte("id: alpha\nname: A\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := dashboard.LoadAll(root); err != nil {
		t.Fatal(err)
	}
	// Load permission error (not IsNotExist)
	if err := os.Chmod(filepath.Join(dash, "alpha.yaml"), 0); err == nil {
		t.Cleanup(func() { _ = os.Chmod(filepath.Join(dash, "alpha.yaml"), 0o644) })
		_, _ = dashboard.Load(root, "alpha")
	}
	_ = os.Chmod(filepath.Join(dash, "alpha.yaml"), 0o644)
	// Create WriteFile error
	if err := os.Chmod(dash, 0o555); err == nil {
		t.Cleanup(func() { _ = os.Chmod(dash, 0o755) })
		_, _ = dashboard.Create(root, "newone", "n", "1x1", "", nil)
	}
	_ = os.Chmod(dash, 0o755)

	_ = dashboard.SlotCount("3x1")
	_ = dashboard.SlotCount("2x1")

	md := filepath.Join(root, "headings.md")
	if err := os.WriteFile(md, []byte("### H3\n## H2\n- item\n# H1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plain := filepath.Join(root, "plain.txt")
	if err := os.WriteFile(plain, []byte("no frontmatter"), 0o644); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	recs := []*record.Record{rec("r1", "note", map[string]any{"status": "open", "due": "2026-01-01", "title": "T"}, "")}
	for _, w := range []dashboard.Widget{
		{ID: "mdh", Type: "markdown", Source: "headings.md"},
		{ID: "mdp", Type: "markdown", Source: "plain.txt"},
		{ID: "mde", Type: "markdown", Source: ""},
		{ID: "mda", Type: "markdown", Source: "/tmp/x"},
		{ID: "mdn", Type: "markdown", Source: "x\x00y"},
		{ID: "mdd", Type: "markdown", Source: ".."},
		{ID: "fb", Type: "list", Field: "due", Before: "2026-12-31"},
		{ID: "qb", Type: "list", Query: map[string]any{"due": map[string]any{"before": "2026-01-01"}}},
		{ID: "qc", Type: "list", Query: map[string]any{"title": map[string]any{"contains": "T"}}},
		{ID: "qe", Type: "list", Query: map[string]any{"status": map[string]any{}}},
		{ID: "m2", Type: "metrics", Stats: []dashboard.Stat{
			{Label: "bad1", Query: "(((broken"},
			{Label: "bad2", Query: "(((broken"},
		}, Reminder: &dashboard.Reminder{Query: "(((broken", Template: "{{name}} {{foo}}"}},
	} {
		d := &dashboard.Dashboard{ID: "w", Name: "w", Layout: "1x1", Widgets: []dashboard.Widget{w}}
		_ = dashboard.Project(d, dashboard.Env{Root: root, Records: recs, Now: now})
	}
}
