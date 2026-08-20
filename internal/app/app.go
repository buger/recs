package app

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"crm/internal/board"
	"crm/internal/contextpkg"
	"crm/internal/index"
	"crm/internal/query"
	"crm/internal/record"
	"crm/internal/store"
	"crm/internal/validate"
)

// App is the shared application layer used by CLI and HTTP.
type App struct {
	Store *store.Store
}

// Implements: INT-REQ-260820-JC9M
func Open(root string) (*App, error) {
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		root = wd
	}
	found, err := store.FindRoot(root)
	if err != nil {
		return nil, err
	}
	return &App{Store: &store.Store{Root: found}}, nil
}

// Implements: INT-REQ-260820-JC9M
func OpenOrCWD(root string) *App {
	if root == "" {
		wd, _ := os.Getwd()
		root = wd
	}
	if found, err := store.FindRoot(root); err == nil {
		root = found
	}
	return &App{Store: &store.Store{Root: root}}
}

// Implements: INT-REQ-260820-JC9M
func (a *App) Root() string { return a.Store.Root }

// Implements: SYS-REQ-260820-KJ34 SW-REQ-260820-MQF2
func (a *App) Init() error { return a.Store.Init() }

// Implements: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y INT-REQ-260820-JC9M
func (a *App) Create(typ string, id string, sets map[string]any, body string) (*record.Record, error) {
	rec := &record.Record{Type: typ, ID: id, Fields: map[string]any{"type": typ}, Body: body}
	if id != "" {
		rec.Set("id", id)
	}
	for k, v := range sets {
		rec.Set(k, v)
	}
	if rec.Body == "" {
		title := rec.GetString("title")
		if title == "" {
			title = rec.GetString("name")
		}
		if title == "" {
			title = rec.ID
		}
		rec.Body = "# " + title + "\n"
	}
	if err := a.Store.Create(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// Implements: SYS-REQ-260820-9J7C
func (a *App) Show(id string) (*record.Record, error) { return a.Store.Get(id) }

// Implements: SYS-REQ-260820-9J7C
func (a *App) List(typ string) ([]*record.Record, error) {
	recs, err := a.Store.LoadAll()
	if err != nil {
		return nil, err
	}
	if typ == "" {
		return recs, nil
	}
	var out []*record.Record
	for _, rec := range recs {
		if rec.Type == typ {
			out = append(out, rec)
		}
	}
	return out, nil
}

// Implements: SYS-REQ-260820-ZTC3 SW-REQ-260820-6EVX
func (a *App) Query(expr string) ([]*record.Record, error) {
	recs, err := a.Store.LoadAll()
	if err != nil {
		return nil, err
	}
	return query.Filter(recs, expr)
}

// Implements: SYS-REQ-260820-HJPH SW-REQ-260820-X37F
func (a *App) Search(q string) ([]*record.Record, error) {
	recs, err := a.Store.LoadAll()
	if err != nil {
		return nil, err
	}
	return query.Search(recs, q), nil
}

// Implements: SYS-REQ-260820-2SQZ
func (a *App) Set(id, field string, value any) (*store.PatchResult, error) {
	return a.Store.Patch(id, map[string]any{field: value}, nil, "")
}

// Implements: SYS-REQ-260820-2SQZ
func (a *App) Patch(id string, sets map[string]any, deletes []string, ifVersion string) (*store.PatchResult, error) {
	return a.Store.Patch(id, sets, deletes, ifVersion)
}

// Implements: SYS-REQ-260820-4628 SW-REQ-260820-NBGR
func (a *App) Board(name string, filters map[string]string) (*board.View, error) {
	b, err := board.Load(a.Root(), name)
	if err != nil {
		return nil, err
	}
	recs, err := a.Store.LoadAll()
	if err != nil {
		return nil, err
	}
	var filtered []*record.Record
	for _, rec := range recs {
		if !b.Matches(rec) {
			continue
		}
		ok := true
		for k, v := range filters {
			if !strings.EqualFold(rec.GetString(k), v) && !containsFold(record.StringSlice(rec.Get(k)), v) {
				ok = false
				break
			}
		}
		if ok {
			filtered = append(filtered, rec)
		}
	}
	return b.Project(filtered), nil
}

// Implements: SYS-REQ-260820-4628
func (a *App) ListBoards() ([]*board.Board, error) {
	return board.LoadAll(a.Root())
}

// Implements: SYS-REQ-260820-BVBE
func (a *App) Move(id, boardName, column string) (*record.Record, string, error) {
	b, err := board.Load(a.Root(), boardName)
	if err != nil {
		return nil, "", err
	}
	rec, err := a.Store.Get(id)
	if err != nil {
		return nil, "", err
	}
	oldPath := rec.Path
	if err := b.ApplyMove(rec, column); err != nil {
		return nil, "", err
	}
	rec.Set("updated_at", time.Now().UTC().Format(time.RFC3339))
	if err := a.Store.Write(rec); err != nil {
		return nil, "", err
	}
	if rec.Path != oldPath {
		return nil, "", fmt.Errorf("move relocated file")
	}
	return rec, oldPath, nil
}

type NextAction struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Date     string `json:"date,omitempty"`
	Priority string `json:"priority,omitempty"`
}

// Implements: SYS-REQ-260820-5C9D SW-REQ-260820-ZKCV
func (a *App) Next() ([]NextAction, error) {
	recs, err := a.Store.LoadAll()
	if err != nil {
		return nil, err
	}
	var out []NextAction
	for _, rec := range recs {
		if rec.Type == "task" && strings.EqualFold(rec.GetString("status"), "open") {
			out = append(out, NextAction{
				ID: rec.ID, Title: record.DisplayName(rec),
				Date: rec.GetString("due"), Priority: rec.GetString("priority"),
			})
			continue
		}
		if rec.Get("next_action") == nil {
			continue
		}
		title := rec.GetString("next_action.action")
		date := rec.GetString("next_action.date")
		if title == "" {
			title = rec.GetString("next_action")
		}
		if title == "" {
			title = record.DisplayName(rec)
		}
		out = append(out, NextAction{ID: rec.ID, Title: title, Date: date, Priority: rec.GetString("priority")})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Date != out[j].Date {
			return out[i].Date < out[j].Date
		}
		return priorityRank(out[i].Priority) < priorityRank(out[j].Priority)
	})
	return out, nil
}

type TriageItem struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
	Title  string `json:"title"`
}

// Implements: SYS-REQ-260820-DCG4 SW-REQ-260820-D5WE
func (a *App) Triage() ([]TriageItem, error) {
	recs, err := a.Store.LoadAll()
	if err != nil {
		return nil, err
	}
	var out []TriageItem
	today := time.Now().UTC().Format("2006-01-02")
	for _, rec := range recs {
		title := record.DisplayName(rec)
		if rec.GetString("status") == "inbox" || rec.GetString("triage_status") == "new" {
			out = append(out, TriageItem{ID: rec.ID, Reason: "inbox", Title: title})
		}
		if rec.GetString("title") == "" && rec.GetString("name") == "" {
			out = append(out, TriageItem{ID: rec.ID, Reason: "missing_metadata", Title: title})
		}
		due := rec.GetString("next_action.date")
		if due == "" {
			due = rec.GetString("due")
		}
		if due != "" && due < today {
			out = append(out, TriageItem{ID: rec.ID, Reason: "overdue", Title: title})
		}
		if rec.GetString("health") == "at_risk" || len(record.StringSlice(rec.Get("blockers"))) > 0 {
			out = append(out, TriageItem{ID: rec.ID, Reason: "blocker", Title: title})
		}
	}
	return out, nil
}

// Implements: SYS-REQ-260820-YWV4
func (a *App) Validate() (*validate.Result, error) {
	recs, err := a.Store.LoadAll()
	if err != nil {
		return nil, err
	}
	return validate.Check(a.Root(), recs)
}

// Implements: SYS-REQ-260820-Q8GR
func (a *App) RebuildIndex() (*index.Snapshot, error) {
	recs, err := a.Store.LoadAll()
	if err != nil {
		return nil, err
	}
	return index.Rebuild(a.Root(), recs)
}

// Implements: SYS-REQ-260820-0TQX
func (a *App) Context(id string) (*contextpkg.Bundle, error) {
	seed, err := a.Store.Get(id)
	if err != nil {
		return nil, err
	}
	all, err := a.Store.LoadAll()
	if err != nil {
		return nil, err
	}
	return contextpkg.Assemble(seed, all), nil
}

// Implements: INT-REQ-260820-JC9M
func containsFold(ss []string, want string) bool {
	for _, s := range ss {
		if strings.EqualFold(s, want) {
			return true
		}
	}
	return false
}

// Implements: INT-REQ-260820-JC9M
func priorityRank(p string) int {
	switch strings.ToLower(p) {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium", "med":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}
