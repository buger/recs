package serve

import (
	"embed"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"

	"crm/internal/app"
	"crm/internal/dashboard"
	"crm/internal/record"
	"crm/internal/store"
)

//go:embed static/index.html
var uiFS embed.FS

// Listen serves the local HTTP API, dashboard gallery, and Kanban UI.
// Implements: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7 INT-REQ-260820-AHKR
func Listen(a *app.App, port int) error {
	ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return err
	}
	return http.Serve(ln, Handler(a))
}

// Implements: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7
func Handler(a *app.App) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/records", func(w http.ResponseWriter, r *http.Request) {
		handleRecords(a, w, r)
	})
	mux.HandleFunc("/api/records/", func(w http.ResponseWriter, r *http.Request) {
		handleRecord(a, w, r)
	})
	mux.HandleFunc("/api/boards", func(w http.ResponseWriter, r *http.Request) {
		handleBoards(a, w, r)
	})
	mux.HandleFunc("/api/boards/", func(w http.ResponseWriter, r *http.Request) {
		handleBoard(a, w, r)
	})
	mux.HandleFunc("/api/dashboards", func(w http.ResponseWriter, r *http.Request) {
		handleDashboards(a, w, r)
	})
	mux.HandleFunc("/api/dashboards/", func(w http.ResponseWriter, r *http.Request) {
		handleDashboard(a, w, r)
	})
	mux.HandleFunc("/api/files/", func(w http.ResponseWriter, r *http.Request) {
		handleFiles(a, w, r)
	})
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		handleSearch(a, w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleGallery(w, r)
	})
	return guardOrigin(mux)
}

// Implements: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7 SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func handleRecords(a *app.App, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		typ := r.URL.Query().Get("type")
		recs, err := a.List(typ)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "records": summaries(recs)})
	case http.MethodPost:
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, 400, err)
			return
		}
		typ, _ := body["type"].(string)
		id, _ := body["id"].(string)
		md, _ := body["body"].(string)
		delete(body, "type")
		delete(body, "id")
		delete(body, "body")
		rec, err := a.Create(typ, id, body, md)
		if err != nil {
			writeErr(w, 400, err)
			return
		}
		writeJSON(w, 201, map[string]any{"ok": true, "record": rec.ID})
	default:
		w.WriteHeader(405)
	}
}

// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA INT-REQ-260821-MRGW
func handleRecord(a *app.App, w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/records/")
	if id == "" {
		w.WriteHeader(404)
		return
	}
	switch r.Method {
	case http.MethodGet:
		view, err := a.View(id)
		if err != nil {
			writeErr(w, 404, err)
			return
		}
		writeJSON(w, 200, map[string]any{
			"ok": true, "record": view.Public, "html": view.HTML,
			"relations": view.Relations, "backlinks": view.Backlinks,
			"attachments": view.Attachments,
		})
	case http.MethodPatch:
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, 400, err)
			return
		}
		ifv, _ := body["_if_version"].(string)
		delete(body, "_if_version")
		var bp *string
		if raw, ok := body["body"]; ok {
			s := fmtSprint(raw)
			bp = &s
			delete(body, "body")
		}
		if bp != nil {
			res, err := a.Edit(id, body, bp, ifv)
			if err != nil {
				writeErr(w, 409, err)
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true, "record": res.Record.ID, "changed": res.Changed})
			return
		}
		res, err := a.Patch(id, body, nil, ifv)
		if err != nil {
			writeErr(w, 409, err)
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "record": res.Record.ID, "changed": res.Changed})
	default:
		w.WriteHeader(405)
	}
}

// Implements: SYS-REQ-260820-4628 SW-REQ-260820-NBGR SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func handleBoards(a *app.App, w http.ResponseWriter, r *http.Request) {
	boards, err := a.ListBoards()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	var names []map[string]string
	for _, b := range boards {
		names = append(names, map[string]string{"id": b.ID, "name": b.Name})
	}
	writeJSON(w, 200, map[string]any{"ok": true, "boards": names})
}

// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func handleBoard(a *app.App, w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/boards/")
	if strings.HasSuffix(rest, "/move") && r.Method == http.MethodPost {
		name := strings.TrimSuffix(rest, "/move")
		name = strings.TrimSuffix(name, "/")
		var body struct {
			ID     string `json:"id"`
			Column string `json:"column"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, 400, err)
			return
		}
		rec, _, err := a.Move(body.ID, name, body.Column)
		if err != nil {
			writeErr(w, 400, err)
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "record": rec.ID})
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	filters := map[string]string{}
	for k, vs := range r.URL.Query() {
		if k != "" && len(vs) > 0 && vs[0] != "" {
			filters[k] = vs[0]
		}
	}
	view, err := a.Board(rest, filters)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	cols := []map[string]any{}
	for _, c := range view.Columns {
		cols = append(cols, map[string]any{"id": c.Column.ID, "title": c.Column.Title, "records": summaries(c.Records)})
	}
	controls := []map[string]string{}
	for _, fc := range view.Board.FilterControls {
		controls = append(controls, map[string]string{"field": fc.Field, "type": fc.Type})
	}
	writeJSON(w, 200, map[string]any{"ok": true, "board": view.Board.ID, "name": view.Board.Name, "columns": cols, "filter_controls": controls, "filters": filters})
}

// Implements: SYS-REQ-260820-456X SW-REQ-260820-EJVT INT-REQ-260820-NHBY
func handleDashboards(a *app.App, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		dashboards, err := a.ListDashboards()
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		var items []map[string]any
		for _, d := range dashboards {
			items = append(items, map[string]any{
				"id": d.ID, "name": d.Name, "description": d.Description,
				"layout": d.Layout, "widgets": len(d.Widgets),
			})
		}
		if items == nil {
			items = []map[string]any{}
		}
		writeJSON(w, 200, map[string]any{"ok": true, "dashboards": items})
	case http.MethodPost:
		var body struct {
			ID          string             `json:"id"`
			Name        string             `json:"name"`
			Layout      string             `json:"layout"`
			Description string             `json:"description"`
			Widgets     []dashboard.Widget `json:"widgets"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, 400, err)
			return
		}
		d, err := a.CreateDashboard(body.ID, body.Name, body.Layout, body.Description, body.Widgets)
		if err != nil {
			writeErr(w, 400, err)
			return
		}
		writeJSON(w, 201, map[string]any{"ok": true, "dashboard": d.ID, "path": d.File})
	default:
		w.WriteHeader(405)
	}
}

// Implements: SYS-REQ-260820-456X SW-REQ-260820-EJVT INT-REQ-260820-NHBY
func handleDashboard(a *app.App, w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/dashboards/")
	if id == "" {
		w.WriteHeader(404)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		return
	}
	proj, err := a.ProjectDashboard(id)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "dashboard": proj})
}

// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func handleFiles(a *app.App, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(405)
		return
	}
	rel := strings.TrimPrefix(r.URL.Path, "/api/files/")
	path, err := a.AttachmentFile(rel)
	if err != nil {
		writeErr(w, 404, err)
		return
	}
	http.ServeFile(w, r, path)
}

// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func handleSearch(a *app.App, w http.ResponseWriter, r *http.Request) {
	recs, err := a.Search(r.URL.Query().Get("q"))
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "records": summaries(recs)})
}

// Implements: SYS-REQ-260820-456X SW-REQ-260820-EJVT SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7
func handleGallery(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data, _ := uiFS.ReadFile("static/index.html") //mcdc:ignore:defensive embedded index.html is always present
	_, _ = w.Write(data)
}

// Implements: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7 INT-REQ-260820-AHKR
func guardOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			if origin := r.Header.Get("Origin"); origin != "" && !allowedOrigin(origin) {
				writeJSON(w, 403, map[string]any{"ok": false, "error": "origin_denied"})
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Implements: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7 INT-REQ-260820-AHKR
func allowedOrigin(origin string) bool {
	switch strings.ToLower(origin) {
	case "http://127.0.0.1", "http://localhost", "https://127.0.0.1", "https://localhost":
		return true
	}
	if strings.HasPrefix(strings.ToLower(origin), "http://127.0.0.1:") || strings.HasPrefix(strings.ToLower(origin), "http://localhost:") {
		return true
	}
	return false
}

// Implements: SYS-REQ-260820-9W1S
func summaries(recs []*record.Record) []map[string]any {
	out := make([]map[string]any, 0, len(recs))
	for _, rec := range recs {
		out = append(out, map[string]any{"id": rec.ID, "type": rec.Type, "title": record.DisplayName(rec), "status": rec.GetString("status")})
	}
	return out
}

// Implements: SYS-REQ-260820-9W1S
func public(rec *record.Record) map[string]any {
	out := map[string]any{"id": rec.ID, "type": rec.Type, "path": rec.Path, "body": rec.Body, "version": rec.Version()}
	for k, v := range rec.Fields {
		out[k] = v
	}
	return out
}

// Implements: SYS-REQ-260820-9W1S
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v) //mcdc:ignore:defensive encoding a map to ResponseWriter is not a product decision
}

// Implements: SYS-REQ-260820-9W1S
func writeErr(w http.ResponseWriter, code int, err error) {
	payload := map[string]any{"ok": false, "error": err.Error()}
	var enumErr *store.EnumError
	if errors.As(err, &enumErr) {
		payload = map[string]any{"ok": false, "error": "invalid_enum", "field": enumErr.Field, "value": enumErr.Value, "allowed": enumErr.Allowed}
	}
	var confErr *store.ConflictError
	if errors.As(err, &confErr) {
		payload = map[string]any{"ok": false, "error": "conflict", "expected_version": confErr.Expected, "current_version": confErr.Current}
	}
	writeJSON(w, code, payload)
}

// Implements: SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func fmtSprint(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
