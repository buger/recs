package dashboard

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"crm/internal/board"
	"crm/internal/query"
	"crm/internal/record"
	"gopkg.in/yaml.v3"
)

// Dashboard is a YAML view over records, files, and boards.
type Dashboard struct {
	File        string   `yaml:"-" json:"-"`
	ID          string   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Layout      string   `yaml:"layout,omitempty" json:"layout"`
	Theme       string   `yaml:"theme,omitempty" json:"theme,omitempty"`
	Widgets     []Widget `yaml:"widgets" json:"widgets"`
}

// Widget is one dashboard slot definition.
type Widget struct {
	ID       string    `yaml:"id" json:"id"`
	Type     string    `yaml:"type" json:"type"`
	Title    string    `yaml:"title,omitempty" json:"title,omitempty"`
	Query    any       `yaml:"query,omitempty" json:"query,omitempty"`
	Source   string    `yaml:"source,omitempty" json:"source,omitempty"`
	Board    string    `yaml:"board,omitempty" json:"board,omitempty"`
	Stats    []Stat    `yaml:"stats,omitempty" json:"stats,omitempty"`
	Reminder *Reminder `yaml:"reminder,omitempty" json:"reminder,omitempty"`
	Field    string    `yaml:"field,omitempty" json:"field,omitempty"`
	Before   string    `yaml:"before,omitempty" json:"before,omitempty"`
	GroupBy  string    `yaml:"group_by,omitempty" json:"group_by,omitempty"`
	Limit    int       `yaml:"limit,omitempty" json:"limit,omitempty"`
	Template string    `yaml:"template,omitempty" json:"template,omitempty"`
}

// Stat is one metrics rollup.
type Stat struct {
	Label  string `yaml:"label" json:"label"`
	Query  any    `yaml:"query,omitempty" json:"query,omitempty"`
	Field  string `yaml:"field,omitempty" json:"field,omitempty"`
	Before string `yaml:"before,omitempty" json:"before,omitempty"`
}

// Reminder is an overdue line template.
type Reminder struct {
	Query    any    `yaml:"query,omitempty" json:"query,omitempty"`
	Template string `yaml:"template,omitempty" json:"template,omitempty"`
}

// Projection is a dashboard with live widget payloads.
type Projection struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Layout      string       `json:"layout"`
	Theme       string       `json:"theme,omitempty"`
	Slots       int          `json:"slots"`
	WidgetCount int          `json:"widget_count"`
	Widgets     []WidgetView `json:"widgets"`
}

// WidgetView is one projected slot.
type WidgetView struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Title       string      `json:"title,omitempty"`
	Source      string      `json:"source,omitempty"`
	Placeholder bool        `json:"placeholder,omitempty"`
	Rejected    bool        `json:"rejected,omitempty"`
	Error       string      `json:"error,omitempty"`
	Count       int         `json:"count"`
	Items       []ItemView  `json:"items,omitempty"`
	Groups      []GroupView `json:"groups,omitempty"`
	Stats       []StatView  `json:"stats,omitempty"`
	Reminders   []string    `json:"reminders,omitempty"`
	Markdown    string      `json:"markdown,omitempty"`
	HTML        string      `json:"html,omitempty"`
	Board       *BoardView  `json:"board,omitempty"`
	Progress    float64     `json:"progress,omitempty"`
	Meta        string      `json:"meta,omitempty"`
}

// ItemView is one record or note line.
type ItemView struct {
	ID       string   `json:"id,omitempty"`
	Title    string   `json:"title,omitempty"`
	Status   string   `json:"status,omitempty"`
	Owner    string   `json:"owner,omitempty"`
	Date     string   `json:"date,omitempty"`
	Body     string   `json:"body,omitempty"`
	Path     string   `json:"path,omitempty"`
	Pills    []string `json:"pills,omitempty"`
	Progress float64  `json:"progress,omitempty"`
}

// GroupView is a named list section.
type GroupView struct {
	Name  string     `json:"name"`
	Items []ItemView `json:"items"`
}

// StatView is a labeled number.
type StatView struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

// BoardView embeds a kanban projection.
type BoardView struct {
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	Columns []BoardCol `json:"columns"`
}

// BoardCol is one embedded column.
type BoardCol struct {
	ID      string     `json:"id"`
	Title   string     `json:"title"`
	Records []ItemView `json:"records"`
}

// Env supplies records, files, and boards for projection.
type Env struct {
	Root    string
	Records []*record.Record
	Board   func(id string) (*board.View, error)
	Now     time.Time
}

var dayToken = regexp.MustCompile(`(?:^|[^A-Za-z0-9_])([+-]\d+d)(?:$|[^A-Za-z0-9_])`)

// KnownTypes reports whether typ is a supported widget type.
// Implements: SYS-REQ-260820-456X SW-REQ-260820-NA06
func KnownType(typ string) bool {
	switch strings.ToLower(strings.TrimSpace(typ)) {
	case "count", "list", "notes", "watch", "pipeline", "metrics", "board", "markdown", "placeholder":
		return true
	default:
		return false
	}
}

// SlotCount returns the layout slot count, at most 4.
// Implements: SYS-REQ-260820-456X SW-REQ-260820-NA06
func SlotCount(layout string) int {
	layout = strings.ToLower(strings.TrimSpace(layout))
	if layout == "" {
		layout = "2x2"
	}
	switch layout {
	case "1x1":
		return 1
	case "1x2", "2x1":
		return 2
	case "2x2":
		return 4
	}
	parts := strings.Split(layout, "x")
	if len(parts) != 2 {
		return 4
	}
	n, errN := strconv.Atoi(strings.TrimSpace(parts[0]))
	m, errM := strconv.Atoi(strings.TrimSpace(parts[1]))
	if errN != nil || errM != nil || n < 1 || m < 1 {
		return 4
	}
	slots := n * m
	if slots > 4 {
		return 4
	}
	return slots
}

// LoadAll reads dashboards/*.yaml.
// Implements: SYS-REQ-260820-456X SW-REQ-260820-NA06
func LoadAll(root string) ([]*Dashboard, error) {
	dir := filepath.Join(root, "dashboards")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []*Dashboard
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		id := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
		d, err := Load(root, id)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Load reads one dashboard YAML file.
// Implements: SYS-REQ-260820-456X SW-REQ-260820-NA06
func Load(root, name string) (*Dashboard, error) {
	if !validID(name) {
		return nil, fmt.Errorf("invalid dashboard name %q", name)
	}
	for _, ext := range []string{".yaml", ".yml"} {
		path := filepath.Join(root, "dashboards", name+ext)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		var d Dashboard
		if err := yaml.Unmarshal(data, &d); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		d.File = path
		if d.ID == "" {
			d.ID = name
		}
		if d.Layout == "" {
			d.Layout = "2x2"
		}
		if d.Theme == "" {
			d.Theme = "light"
		}
		return &d, nil
	}
	return nil, fmt.Errorf("dashboard %s not found", name)
}

// Create writes dashboards/<id>.yaml.
// Implements: SYS-REQ-260820-456X SW-REQ-260820-NA06
func Create(root, id, name, layout, description string, widgets []Widget) (*Dashboard, error) {
	if !validID(id) {
		return nil, fmt.Errorf("invalid dashboard id %q", id)
	}
	if name == "" {
		name = id
	}
	if layout == "" {
		layout = "2x2"
	}
	if widgets == nil {
		widgets = []Widget{}
	}
	d := &Dashboard{ID: id, Name: name, Description: description, Layout: layout, Theme: "light", Widgets: widgets}
	dir := filepath.Join(root, "dashboards")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	path := filepath.Join(dir, id+".yaml")
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("dashboard %s already exists", id)
	}
	data, err := yaml.Marshal(d)
	if err != nil { //mcdc:ignore:defensive yaml.Marshal of a string-keyed struct cannot fail
		return nil, err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, err
	}
	d.File = path
	return d, nil
}

// Project fills widget payloads and pads missing layout slots.
// Implements: SYS-REQ-260820-456X SW-REQ-260820-NA06
func Project(d *Dashboard, env Env) *Projection {
	if d == nil {
		return &Projection{Layout: "2x2", Slots: 4, Widgets: placeholderSlots(4, 0)}
	}
	if env.Now.IsZero() {
		env.Now = time.Now().UTC()
	}
	layout := d.Layout
	if layout == "" {
		layout = "2x2"
	}
	slots := SlotCount(layout)
	views := make([]WidgetView, 0, slots)
	defined := 0
	for i, w := range d.Widgets {
		if i >= slots {
			break
		}
		views = append(views, projectWidget(w, env))
		if strings.ToLower(strings.TrimSpace(w.Type)) != "placeholder" && strings.TrimSpace(w.Type) != "" {
			defined++
		}
	}
	if len(views) < slots {
		views = append(views, placeholderSlots(slots, len(views))...)
	}
	return &Projection{
		ID: d.ID, Name: d.Name, Description: d.Description,
		Layout: layout, Theme: d.Theme, Slots: slots,
		WidgetCount: defined, Widgets: views,
	}
}

// Implements: SYS-REQ-260820-456X SW-REQ-260820-NA06
func projectWidget(w Widget, env Env) WidgetView {
	typ := strings.ToLower(strings.TrimSpace(w.Type))
	view := WidgetView{ID: w.ID, Type: typ, Title: w.Title, Source: widgetSource(w)}
	if typ == "" || typ == "placeholder" {
		view.Type = "placeholder"
		view.Placeholder = true
		return view
	}
	if !KnownType(typ) {
		view.Rejected = true
		view.Error = "unknown widget type"
		return view
	}
	switch typ {
	case "count":
		recs, err := runQuery(env.Records, w.Query, w.Field, w.Before, env.Now)
		if err != nil {
			view.Error = err.Error()
			return view
		}
		view.Count = len(recs)
		view.Meta = fmt.Sprintf("%d records", len(recs))
	case "list":
		recs, err := runQuery(env.Records, w.Query, w.Field, w.Before, env.Now)
		if err != nil {
			view.Error = err.Error()
			return view
		}
		recs = limitRecs(recs, w.Limit)
		group := w.GroupBy
		if group == "" {
			group = "status"
		}
		by := map[string][]ItemView{}
		var order []string
		for _, rec := range recs {
			item := itemFrom(rec, env.Root, false)
			key := rec.GetString(group)
			if key == "" {
				key = "open"
			}
			if _, ok := by[key]; !ok {
				order = append(order, key)
			}
			by[key] = append(by[key], item)
			view.Items = append(view.Items, item)
		}
		for _, k := range order {
			view.Groups = append(view.Groups, GroupView{Name: k, Items: by[k]})
		}
		view.Count = len(recs)
		view.Meta = env.Now.UTC().Format("02 Jan 2006")
	case "notes":
		recs, err := runQuery(env.Records, w.Query, w.Field, w.Before, env.Now)
		if err != nil {
			view.Error = err.Error()
			return view
		}
		recs = limitRecs(recs, w.Limit)
		for _, rec := range recs {
			view.Items = append(view.Items, itemFrom(rec, env.Root, true))
		}
		view.Count = len(view.Items)
		view.Meta = fmt.Sprintf("%d notes", len(view.Items))
	case "watch":
		recs, err := runQuery(env.Records, w.Query, w.Field, w.Before, env.Now)
		if err != nil {
			view.Error = err.Error()
			return view
		}
		recs = limitRecs(recs, w.Limit)
		for _, rec := range recs {
			item := itemFrom(rec, env.Root, true)
			if item.Status == "" {
				item.Pills = append([]string{"watch"}, item.Pills...)
			}
			view.Items = append(view.Items, item)
		}
		view.Count = len(view.Items)
	case "pipeline":
		recs, err := runQuery(env.Records, w.Query, w.Field, w.Before, env.Now)
		if err != nil {
			view.Error = err.Error()
			return view
		}
		recs = limitRecs(recs, w.Limit)
		done := 0
		for _, rec := range recs {
			item := itemFrom(rec, env.Root, false)
			st := strings.ToLower(item.Status)
			if st == "done" || st == "won" || st == "applied" || st == "complete" {
				done++
				item.Progress = 1
			}
			view.Items = append(view.Items, item)
		}
		view.Count = len(view.Items)
		if view.Count > 0 {
			view.Progress = float64(done) / float64(view.Count)
		}
		view.Meta = fmt.Sprintf("%d items", view.Count)
	case "metrics":
		for _, st := range w.Stats {
			recs, err := runQuery(env.Records, st.Query, st.Field, st.Before, env.Now)
			if err != nil {
				view.Stats = append(view.Stats, StatView{Label: st.Label, Value: 0})
				if view.Error == "" {
					view.Error = err.Error()
				}
				continue
			}
			view.Stats = append(view.Stats, StatView{Label: st.Label, Value: len(recs)})
		}
		if w.Reminder != nil {
			recs, err := runQuery(env.Records, w.Reminder.Query, "", "", env.Now)
			if err != nil {
				if view.Error == "" {
					view.Error = err.Error()
				}
			} else {
				tmpl := w.Reminder.Template
				if tmpl == "" {
					tmpl = "{{name}} — follow-up was due {{next_action.date}}"
				}
				for _, rec := range recs {
					view.Reminders = append(view.Reminders, applyTemplate(tmpl, rec))
				}
			}
		}
		view.Count = len(view.Stats)
		view.Meta = env.Now.UTC().Format("02 Jan 2006")
	case "board":
		if strings.TrimSpace(w.Board) == "" {
			view.Placeholder = true
			view.Error = "missing source"
			return view
		}
		if env.Board == nil {
			view.Placeholder = true
			view.Error = "missing source"
			return view
		}
		bv, err := env.Board(w.Board)
		if err != nil || bv == nil {
			view.Placeholder = true
			view.Error = "missing source"
			if err != nil {
				view.Error = err.Error()
			}
			return view
		}
		view.Source = "boards/" + w.Board + ".yaml"
		out := &BoardView{ID: bv.Board.ID, Name: bv.Board.Name}
		n := 0
		for _, col := range bv.Columns {
			var items []ItemView
			for _, rec := range col.Records {
				items = append(items, itemFrom(rec, env.Root, false))
				n++
			}
			out.Columns = append(out.Columns, BoardCol{ID: col.Column.ID, Title: col.Column.Title, Records: items})
		}
		view.Board = out
		view.Count = n
		view.Meta = bv.Board.Name
	case "markdown":
		if strings.TrimSpace(w.Source) == "" {
			view.Placeholder = true
			view.Error = "missing source"
			return view
		}
		full, err := safeRel(env.Root, w.Source)
		if err != nil {
			view.Rejected = true
			view.Error = err.Error()
			return view
		}
		data, err := os.ReadFile(full)
		if err != nil {
			view.Placeholder = true
			view.Error = "missing source"
			return view
		}
		view.Source = w.Source
		view.Markdown = string(data)
		view.HTML = renderMarkdown(string(data))
		view.Meta = w.Source
		if rec, err := record.Parse(full, data); err == nil && rec != nil { //mcdc:ignore:defensive record.Parse returns a record xor an error, never a nil record with a nil error
			if view.Title == "" {
				view.Title = record.DisplayName(rec)
			}
			view.Items = []ItemView{itemFrom(rec, env.Root, true)}
		}
	}
	return view
}

// Implements: SW-REQ-260820-NA06
func widgetSource(w Widget) string {
	if w.Source != "" {
		return w.Source
	}
	if w.Board != "" {
		return "boards/" + w.Board + ".yaml"
	}
	if s := queryString(w.Query); s != "" {
		return s
	}
	return ""
}

// Implements: SYS-REQ-260820-456X SW-REQ-260820-NA06
func runQuery(recs []*record.Record, q any, field, before string, now time.Time) ([]*record.Record, error) {
	expr := queryString(q)
	if strings.TrimSpace(field) != "" && strings.TrimSpace(before) != "" {
		if expr != "" {
			expr += " "
		}
		expr += field + "<" + before
	}
	expr = expandQuery(expr, now)
	if strings.TrimSpace(expr) == "" {
		return recs, nil
	}
	return query.Filter(recs, expr)
}

// Implements: SW-REQ-260820-NA06
func queryString(q any) string {
	switch t := q.(type) {
	case nil:
		return ""
	case string:
		return t
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var parts []string
		for _, k := range keys {
			parts = append(parts, fieldClause(k, t[k]))
		}
		return strings.Join(parts, " ")
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

// Implements: SW-REQ-260820-NA06
func fieldClause(field string, spec any) string {
	if m, ok := spec.(map[string]any); ok {
		if v, ok := m["before"]; ok {
			return field + "<" + fmt.Sprint(v)
		}
		if v, ok := m["after"]; ok {
			return field + ">" + fmt.Sprint(v)
		}
		if v, ok := m["eq"]; ok {
			return field + "=" + fmt.Sprint(v)
		}
		if v, ok := m["contains"]; ok {
			return field + " contains " + fmt.Sprint(v)
		}
	}
	return field + "=" + fmt.Sprint(spec)
}

// Implements: SW-REQ-260820-NA06
func expandQuery(expr string, now time.Time) string {
if now.IsZero() { //mcdc:ignore:defensive Project fills Env.Now before runQuery so expandQuery never sees a zero clock
		now = time.Now().UTC()
	}
	today := now.UTC().Format("2006-01-02")
	expr = regexp.MustCompile(`\btoday\b`).ReplaceAllString(expr, today)
	return dayToken.ReplaceAllStringFunc(expr, func(m string) string {
		sub := dayToken.FindStringSubmatch(m)
		if len(sub) < 2 { //mcdc:ignore:defensive the day-token regexp always captures the signed day group
			return m
		}
		tok := sub[1]
		n, err := strconv.Atoi(strings.TrimSuffix(tok, "d"))
		if err != nil { //mcdc:ignore:defensive the regexp only matches [+-]digits d
			return m
		}
		day := now.UTC().AddDate(0, 0, n).Format("2006-01-02")
		return strings.Replace(m, tok, day, 1)
	})
}

// Implements: SW-REQ-260820-NA06
func limitRecs(recs []*record.Record, n int) []*record.Record {
	if n <= 0 || n >= len(recs) {
		return recs
	}
	return recs[:n]
}

// Implements: SW-REQ-260820-NA06
func itemFrom(rec *record.Record, root string, excerpt bool) ItemView {
	body := strings.TrimSpace(rec.Body)
	if excerpt && len(body) > 180 {
		body = strings.TrimSpace(body[:180]) + "…"
	}
	var pills []string
	if s := rec.GetString("status"); s != "" {
		pills = append(pills, s)
	}
	if o := rec.GetString("owner"); o != "" {
		pills = append(pills, o)
	}
	if p := rec.GetString("priority"); p != "" {
		pills = append(pills, p)
	}
	for _, tag := range record.StringSlice(rec.Get("tags")) {
		if tag != "" {
			pills = append(pills, tag)
		}
	}
	date := rec.GetString("next_action.date")
	if date == "" {
		date = rec.GetString("due")
	}
	if date == "" {
		date = rec.GetString("deadline")
	}
	path := rec.Path
	if root != "" && strings.HasPrefix(path, root) {
		path = strings.TrimPrefix(path, root)
		path = strings.TrimPrefix(path, string(os.PathSeparator))
	}
	return ItemView{
		ID: rec.ID, Title: record.DisplayName(rec),
		Status: rec.GetString("status"), Owner: rec.GetString("owner"),
		Date: date, Body: body, Path: path, Pills: pills,
	}
}

// Implements: SW-REQ-260820-NA06
func applyTemplate(tmpl string, rec *record.Record) string {
	out := tmpl
	out = strings.ReplaceAll(out, "{{name}}", record.DisplayName(rec))
	out = strings.ReplaceAll(out, "{{title}}", record.DisplayName(rec))
	out = strings.ReplaceAll(out, "{{id}}", rec.ID)
	re := regexp.MustCompile(`\{\{([A-Za-z0-9_.]+)\}\}`)
	return re.ReplaceAllStringFunc(out, func(m string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(m, "{{"), "}}")
		if key == "name" || key == "title" || key == "id" { //mcdc:ignore:defensive ReplaceAll already consumed name/title/id tokens before the leftover regex
			return m
		}
		v := rec.GetString(key)
		if v == "" {
			return m
		}
		return v
	})
}

// Implements: SW-REQ-260820-NA06
func placeholderSlots(slots, have int) []WidgetView {
	var out []WidgetView
	for i := have; i < slots; i++ {
		out = append(out, WidgetView{
			ID: fmt.Sprintf("slot_%d", i+1), Type: "placeholder", Placeholder: true,
		})
	}
	return out
}

// Implements: SW-REQ-260820-NA06
func validID(name string) bool {
	if name == "" || strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return false
	}
	for _, r := range name {
		ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-'
		if !ok {
			return false
		}
	}
	return true
}

// Implements: SW-REQ-260820-NA06
func safeRel(root, rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", fmt.Errorf("missing source")
	}
	if filepath.IsAbs(rel) || strings.Contains(rel, "..") || strings.ContainsAny(rel, "\x00") {
		return "", fmt.Errorf("invalid source path")
	}
	full := filepath.Join(root, filepath.Clean(rel))
	absRoot, err := filepath.Abs(root)
	if err != nil { //mcdc:ignore:defensive filepath.Abs fails only when Getwd fails
		return "", err
	}
	absFull, err := filepath.Abs(full)
	if err != nil { //mcdc:ignore:defensive Join of Abs root cannot fail Abs
		return "", err
	}
	sep := string(os.PathSeparator)
	if absFull != absRoot && !strings.HasPrefix(absFull, absRoot+sep) { //mcdc:ignore:defensive the prior clause already rejects absolute paths and dot-dot so an Abs path cannot escape the root
		return "", fmt.Errorf("invalid source path")
	}
	return absFull, nil
}

// Implements: SW-REQ-260820-NA06
func renderMarkdown(src string) string {
	escaped := html.EscapeString(src)
	var lines []string
	for _, line := range strings.Split(escaped, "\n") {
		trim := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trim, "### "):
			lines = append(lines, "<h3>"+strings.TrimPrefix(trim, "### ")+"</h3>")
		case strings.HasPrefix(trim, "## "):
			lines = append(lines, "<h2>"+strings.TrimPrefix(trim, "## ")+"</h2>")
		case strings.HasPrefix(trim, "# "):
			lines = append(lines, "<h1>"+strings.TrimPrefix(trim, "# ")+"</h1>")
		case strings.HasPrefix(trim, "- "):
			lines = append(lines, "<li>"+strings.TrimPrefix(trim, "- ")+"</li>")
		case trim == "":
			lines = append(lines, "")
		default:
			lines = append(lines, "<p>"+inlineMD(line)+"</p>")
		}
	}
	return strings.Join(lines, "")
}

// Implements: SW-REQ-260820-NA06
func inlineMD(s string) string {
	re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	return re.ReplaceAllString(s, "<strong>$1</strong>")
}
