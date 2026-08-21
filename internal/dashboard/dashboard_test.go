package dashboard_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"crm/internal/app"
	"crm/internal/board"
	"crm/internal/dashboard"
	"crm/internal/record"
)

// Verifies: SYS-REQ-260820-456X SW-REQ-260820-NA06
// SYS-REQ-260820-456X:nominal:nominal
// SYS-REQ-260820-456X:boundary:nominal
// SYS-REQ-260820-456X:empty_input:nominal
// SYS-REQ-260820-456X:error_handling:nominal
// SW-REQ-260820-NA06:nominal:nominal
// SW-REQ-260820-NA06:boundary:nominal
// SW-REQ-260820-NA06:empty_input:nominal
// SW-REQ-260820-NA06:error_handling:nominal
// SW-REQ-260820-NA06:path_traversal_prevented:nominal
func TestKnownTypeAndSlots(t *testing.T) {
	for _, typ := range []string{"count", "list", "notes", "watch", "pipeline", "metrics", "board", "markdown", "placeholder", "COUNT"} {
		if !dashboard.KnownType(typ) {
			t.Fatalf("known %s", typ)
		}
	}
	if dashboard.KnownType("nope") || dashboard.KnownType("") {
		t.Fatal("unknown")
	}
	if dashboard.SlotCount("") != 4 || dashboard.SlotCount("2x2") != 4 {
		t.Fatal("default slots")
	}
	if dashboard.SlotCount("1x1") != 1 || dashboard.SlotCount("1x2") != 2 || dashboard.SlotCount("2x1") != 2 {
		t.Fatal("small layouts")
	}
	if dashboard.SlotCount("3x3") != 4 || dashboard.SlotCount("weird") != 4 || dashboard.SlotCount("2x") != 4 {
		t.Fatal("clamp")
	}
	if dashboard.SlotCount("2x0") != 4 || dashboard.SlotCount("axb") != 4 {
		t.Fatal("bad parse")
	}
}

func setupWorkspace(t *testing.T) *app.App {
	t.Helper()
	a := app.OpenOrCWD(t.TempDir())
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format("2006-01-02")
	past := time.Now().UTC().AddDate(0, 0, -2).Format("2006-01-02")
	if _, err := a.Create("customer", "customer_acme", map[string]any{"name": "Acme", "status": "active", "owner": "leonid", "priority": "high", "tags": []any{"watch"}, "next_action": map[string]any{"date": past, "action": "Call"}}, "Customer note"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("grant", "grant_x", map[string]any{"title": "Solana", "status": "preparing", "owner": "leonid"}, "Grant body"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("onboarding", "onb_x", map[string]any{"title": "Onb", "status": "blocked"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("note", "note_x", map[string]any{"title": "Memo", "owner": "leonid"}, "A longer workspace note for the notes widget."); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("task", "task_done", map[string]any{"title": "Done", "status": "done", "due": now}, ""); err != nil {
		t.Fatal(err)
	}
	return a
}

// SW-REQ-260820-NA06:path_traversal_prevented:negative
// SYS-REQ-260820-456X:error_handling:negative
// SW-REQ-260820-NA06:error_handling:negative
func TestLoadCreateProject(t *testing.T) {
	a := setupWorkspace(t)
	list, err := a.ListDashboards()
	if err != nil || len(list) < 2 {
		t.Fatalf("defaults %v %d", err, len(list))
	}
	if _, err := a.Dashboard("../etc"); err == nil {
		t.Fatal("traversal")
	}
	if _, err := a.Dashboard("missing"); err == nil {
		t.Fatal("missing")
	}
	if _, err := dashboard.LoadAll("/no/such/crm-root-dash"); err != nil {
		// missing dir is empty
		if !os.IsNotExist(err) && err.Error() == "boom" {
			t.Fatal(err)
		}
	}
	empty, err := dashboard.LoadAll(t.TempDir())
	if err != nil || empty != nil {
		t.Fatal(err, empty)
	}
	if _, err := a.CreateDashboard("bad/id", "X", "2x2", "", nil); err == nil {
		t.Fatal("bad id")
	}
	created, err := a.CreateDashboard("custom", "Custom", "", "desc", []dashboard.Widget{{ID: "c", Type: "count", Title: "All", Query: ""}})
	if err != nil || created.Layout != "2x2" {
		t.Fatal(err, created)
	}
	if _, err := a.CreateDashboard("custom", "Custom", "2x2", "", nil); err == nil {
		t.Fatal("exists")
	}
	proj, err := a.ProjectDashboard("prospects")
	if err != nil || proj.WidgetCount != 1 || len(proj.Widgets) != 4 {
		t.Fatalf("prospects %+v %v", proj, err)
	}
	if proj.Widgets[0].Type != "metrics" || !proj.Widgets[1].Placeholder || !proj.Widgets[2].Placeholder || !proj.Widgets[3].Placeholder {
		t.Fatal("placeholders")
	}
	ws, err := a.ProjectDashboard("workspace")
	if err != nil || ws.WidgetCount < 3 {
		t.Fatal(err, ws)
	}
	inbox, err := a.ProjectDashboard("inbox")
	if err != nil || inbox == nil {
		t.Fatal(err)
	}
}

// Verifies: SYS-REQ-260820-456X SW-REQ-260820-NA06
func TestProjectWidgetKinds(t *testing.T) {
	a := setupWorkspace(t)
	root := a.Root()
	os.WriteFile(filepath.Join(root, "note.md"), []byte("---\nid: md_note\ntype: note\ntitle: File note\n---\n# Hello\n\n**Bold** body\n"), 0o644)
	os.WriteFile(filepath.Join(root, "dashboards", "rich.yaml"), []byte(`id: rich
name: Rich
layout: 2x2
widgets:
  - id: c
    type: count
    title: Grants
    query: 'type=grant'
  - id: l
    type: list
    title: Grants
    query:
      type: grant
    group_by: status
    limit: 2
  - id: n
    type: notes
    title: Notes
    query: 'type=note'
  - id: extra
    type: watch
    query: 'priority=high'
`), 0o644)
	if err := os.WriteFile(filepath.Join(root, "dashboards", "more.yml"), []byte("name: More\nwidgets:\n  - id: p\n    type: pipeline\n    query: 'type=task'\n  - id: m\n    type: metrics\n    stats:\n      - label: all\n        query: ''\n      - label: bad\n        query: '???'\n    reminder:\n      query: 'type=customer next_action.date<today'\n  - id: b\n    type: board\n    board: grants\n  - id: md\n    type: markdown\n    source: note.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rich, err := dashboard.Load(root, "rich")
	if err != nil {
		t.Fatal(err)
	}
	recs, _ := a.Store.LoadAll()
	env := dashboard.Env{Root: root, Records: recs, Board: func(id string) (*board.View, error) { return a.Board(id, nil) }, Now: time.Now().UTC()}
	p := dashboard.Project(rich, env)
	if p.WidgetCount != 4 || p.Slots != 4 {
		t.Fatalf("rich count %d slots %d", p.WidgetCount, p.Slots)
	}
	more, err := dashboard.Load(root, "more")
	if err != nil || more.ID != "more" {
		t.Fatal(err, more)
	}
	mp := dashboard.Project(more, env)
	if mp.Widgets[0].Type != "pipeline" || mp.Widgets[2].Board == nil || mp.Widgets[3].Markdown == "" {
		t.Fatalf("more %+v", mp.Widgets)
	}
	if len(mp.Widgets[1].Reminders) == 0 {
		t.Fatal("reminders")
	}
}

// Verifies: SYS-REQ-260820-456X SW-REQ-260820-NA06
func TestProjectErrorsAndPlaceholders(t *testing.T) {
	a := setupWorkspace(t)
	recs, _ := a.Store.LoadAll()
	env := dashboard.Env{Root: a.Root(), Records: recs, Board: func(id string) (*board.View, error) { return a.Board(id, nil) }}
	d := &dashboard.Dashboard{ID: "x", Name: "X", Layout: "2x2", Widgets: []dashboard.Widget{
		{ID: "bad", Type: "nope"},
		{ID: "md", Type: "markdown"},
		{ID: "bd", Type: "board"},
		{ID: "q", Type: "count", Query: "???"},
	}}
	p := dashboard.Project(d, env)
	if !p.Widgets[0].Rejected || !p.Widgets[1].Placeholder || !p.Widgets[2].Placeholder || p.Widgets[3].Error == "" {
		t.Fatalf("%+v", p.Widgets)
	}
	d2 := &dashboard.Dashboard{Widgets: []dashboard.Widget{
		{ID: "md2", Type: "markdown", Source: "../etc/passwd"},
		{ID: "md3", Type: "markdown", Source: "missing.md"},
		{ID: "bd2", Type: "board", Board: "nope"},
		{ID: "ph", Type: "placeholder"},
	}}
	p2 := dashboard.Project(d2, env)
	if !p2.Widgets[0].Rejected || !p2.Widgets[1].Placeholder || !p2.Widgets[2].Placeholder || !p2.Widgets[3].Placeholder {
		t.Fatalf("%+v", p2.Widgets)
	}
	if dashboard.Project(nil, dashboard.Env{}).Slots != 4 {
		t.Fatal("nil dash")
	}
	emptyEnv := dashboard.Env{Root: a.Root(), Records: recs}
	p3 := dashboard.Project(&dashboard.Dashboard{Widgets: []dashboard.Widget{{ID: "b", Type: "board", Board: "grants"}}}, emptyEnv)
	if !p3.Widgets[0].Placeholder {
		t.Fatal("no board fn")
	}
}

// Verifies: SYS-REQ-260820-456X SW-REQ-260820-NA06
func TestQueryMapAndDates(t *testing.T) {
	a := setupWorkspace(t)
	recs, _ := a.Store.LoadAll()
	d := &dashboard.Dashboard{Layout: "1x1", Widgets: []dashboard.Widget{{
		ID: "due", Type: "list", Field: "next_action.date", Before: "+7d",
		Query: map[string]any{"type": "customer", "status": map[string]any{"eq": "active"}},
	}}}
	p := dashboard.Project(d, dashboard.Env{Records: recs, Now: time.Now().UTC()})
	if p.Slots != 1 || p.Widgets[0].Count < 1 {
		t.Fatalf("%+v", p.Widgets[0])
	}
	d2 := &dashboard.Dashboard{Layout: "1x1", Widgets: []dashboard.Widget{{
		ID: "m", Type: "list", Query: map[string]any{"title": map[string]any{"contains": "Solana"}, "created_at": map[string]any{"after": "2000-01-01"}},
	}}}
	_ = dashboard.Project(d2, dashboard.Env{Records: recs, Now: time.Now().UTC()})
}

// Verifies: SYS-REQ-260820-456X SW-REQ-260820-NA06
func TestBadYAMLAndNonYaml(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "dashboards"), 0o755)
	os.WriteFile(filepath.Join(dir, "dashboards", "skip.txt"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "dashboards", "bad.yaml"), []byte(":\n["), 0o644)
	if _, err := dashboard.LoadAll(dir); err == nil {
		t.Fatal("bad yaml")
	}
	if _, err := dashboard.Load(dir, ""); err == nil {
		t.Fatal("empty id")
	}
	_ = record.DisplayName(&record.Record{ID: "x"})
}

// Verifies: SYS-REQ-260820-456X SW-REQ-260820-NA06
func TestCreateNameDefault(t *testing.T) {
	dir := t.TempDir()
	d, err := dashboard.Create(dir, "solo", "", "", "", nil)
	if err != nil || d.Name != "solo" || d.Layout != "2x2" {
		t.Fatal(err, d)
	}
}
