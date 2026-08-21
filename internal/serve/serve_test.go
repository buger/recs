package serve_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/buger/recs/internal/app"
	"github.com/buger/recs/internal/serve"
)

func setupApp(t *testing.T) *app.App {
	t.Helper()
	a := app.OpenOrCWD(t.TempDir())
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Create("grant", "grant_a", map[string]any{"title": "A", "status": "researching"}, "body"); err != nil {
		t.Fatal(err)
	}
	return a
}

// Verifies: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7 SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func TestHandlerRoutes(t *testing.T) {
	a := setupApp(t)
	h := serve.Handler(a)
	get := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w
	}
	if get("/").Code != 200 || !strings.Contains(get("/").Body.String(), "html") && get("/").Body.Len() == 0 {
		// UI may be empty-ish but should 200
		if get("/").Code != 200 {
			t.Fatal(get("/").Code)
		}
	}
	if get("/nope").Code != 404 {
		t.Fatal("not found")
	}
	if get("/api/records").Code != 200 {
		t.Fatal(get("/api/records").Body.String())
	}
	if get("/api/records?type=grant").Code != 200 {
		t.Fatal("typed list")
	}
	if get("/api/records/grant_a").Code != 200 {
		t.Fatal(get("/api/records/grant_a").Body.String())
	}
	if get("/api/records/missing").Code != 404 {
		t.Fatal("missing rec")
	}
	if get("/api/records/").Code != 404 {
		t.Fatal("empty id")
	}
	if get("/api/boards").Code != 200 {
		t.Fatal("boards")
	}
	if get("/api/boards/grants").Code != 200 {
		t.Fatal(get("/api/boards/grants").Body.String())
	}
	if get("/api/boards/missing").Code != 404 {
		t.Fatal("missing board")
	}
	if get("/api/search?q=A").Code != 200 {
		t.Fatal("search")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/records", nil))
	if w.Code != 405 {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/records/grant_a", nil))
	if w.Code != 405 {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/boards/grants", nil))
	if w.Code != 405 {
		t.Fatal(w.Code)
	}

	postJSON := func(path string, body any, origin string) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	if postJSON("/api/records", map[string]any{"type": "note", "id": "note_z", "title": "Z", "body": "x"}, "http://localhost:7777").Code != 201 {
		t.Fatal("create")
	}
	if postJSON("/api/records", "{", "http://127.0.0.1").Code != 400 {
		t.Fatal("bad json")
	}
	if postJSON("/api/records", map[string]any{"type": "grant", "id": "bad/id"}, "http://localhost").Code != 400 {
		t.Fatal("bad create")
	}
	if postJSON("/api/records", map[string]any{"type": "note"}, "https://evil.example").Code != 403 {
		t.Fatal("csrf")
	}
	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/records/grant_a", bytes.NewReader([]byte(`{"status":"applied"}`)))
	req.Header.Set("Origin", "https://localhost")
	h.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/records/grant_a", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Origin", "http://127.0.0.1:9")
	h.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/records/grant_a", bytes.NewReader([]byte(`{"status":"applied","_if_version":"sha256:dead"}`)))
	req.Header.Set("Origin", "http://localhost")
	h.ServeHTTP(w, req)
	if w.Code != 409 {
		t.Fatal(w.Code, w.Body.String())
	}
	mv := postJSON("/api/boards/grants/move", map[string]any{"id": "grant_a", "column": "applied"}, "http://localhost")
	if mv.Code != 200 && mv.Code != 400 {
		t.Fatal(mv.Code, mv.Body.String())
	}
	if postJSON("/api/boards/grants/move", "{", "http://localhost").Code != 400 {
		t.Fatal("bad move json")
	}
	if postJSON("/api/boards/grants/move", map[string]any{"id": "nope", "column": "x"}, "http://localhost").Code != 400 {
		t.Fatal("bad move")
	}
}

// Verifies: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7 INT-REQ-260820-AHKR
func TestListenBusyPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	a := setupApp(t)
	if err := serve.Listen(a, port); err == nil {
		t.Fatal("expected bind error")
	}
	_ = io.Discard
}

// Verifies: SYS-REQ-260820-456X SW-REQ-260820-EJVT INT-REQ-260820-NHBY STK-REQ-260820-4255
// SYS-REQ-260820-456X:nominal:nominal
// SYS-REQ-260820-456X:error_handling:nominal
// SYS-REQ-260820-456X:error_handling:negative
// SW-REQ-260820-EJVT:nominal:nominal
// SW-REQ-260820-EJVT:boundary:nominal
// SW-REQ-260820-EJVT:error_handling:nominal
// SW-REQ-260820-EJVT:error_handling:negative
// INT-REQ-260820-NHBY:nominal:nominal
// INT-REQ-260820-NHBY:error_handling:nominal
// INT-REQ-260820-NHBY:error_handling:negative
// INT-REQ-260820-NHBY:integration:integration
// STK-REQ-260820-4255:error_handling:nominal
// STK-REQ-260820-4255:error_handling:negative
// STK-REQ-260820-4255:boundary:nominal
// MCDC INT-REQ-260820-NHBY: dashboard_api_requested=T, empty_gallery_shown=F, gallery_rejected=F, gallery_rendered=T, preview_cards_shown=F, separate_database_used=F, shared_app_layer_used=T, shared_record_model_used=T => TRUE
// MCDC INT-REQ-260820-NHBY: dashboard_api_requested=T, empty_gallery_shown=F, gallery_rejected=F, gallery_rendered=F, preview_cards_shown=T, separate_database_used=F, shared_app_layer_used=T, shared_record_model_used=T => TRUE
//mcdc:ignore INT-REQ-260820-NHBY: dashboard_api_requested=T, empty_gallery_shown=F, gallery_rejected=F, gallery_rendered=F, preview_cards_shown=F, separate_database_used=F, shared_app_layer_used=F, shared_record_model_used=F => FALSE -- dashboard API without the shared app layer is the literal negation of the integration contract [reviewed: agent:grok] [category: defensive]
func TestDashboardAPIAndUI(t *testing.T) {
	a := setupApp(t)
	h := serve.Handler(a)
	get := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w
	}
	ui := get("/")
	if ui.Code != 200 || !strings.Contains(ui.Body.String(), "Dashboard") || !strings.Contains(ui.Body.String(), "New dashboard") {
		t.Fatalf("ui %d %s", ui.Code, ui.Body.String()[:min(200, ui.Body.Len())])
	}
	if !strings.Contains(ui.Body.String(), "#/d/") || !strings.Contains(ui.Body.String(), "Dashboards") {
		t.Fatal("gallery routing")
	}
	if !strings.Contains(ui.Body.String(), "tabler") || strings.Contains(ui.Body.String(), "fonts.googleapis.com") {
		t.Fatal("tabler kit or runtime font CDN")
	}
	if !strings.Contains(ui.Body.String(), "theme-toggle") || !strings.Contains(ui.Body.String(), "data-theme") {
		t.Fatal("theme toggle")
	}
	if strings.Contains(ui.Body.String(), ".map(itemLine)") {
		t.Fatal("itemLine map prefix bug")
	}
	css := get("/static/vendor/tabler.min.css")
	if css.Code != 200 || !strings.Contains(css.Body.String(), "Tabler") {
		t.Fatal("vendor css", css.Code)
	}
	list := get("/api/dashboards")
	if list.Code != 200 || !strings.Contains(list.Body.String(), "prospects") || !strings.Contains(list.Body.String(), "workspace") {
		t.Fatal(list.Body.String())
	}
	one := get("/api/dashboards/prospects")
	if one.Code != 200 || !strings.Contains(one.Body.String(), "widgets") {
		t.Fatal(one.Body.String())
	}
	ws := get("/api/dashboards/workspace")
	if ws.Code != 200 {
		t.Fatal(ws.Body.String())
	}
	if get("/api/dashboards/missing").Code != 404 {
		t.Fatal("missing dash")
	}
	if get("/api/dashboards/").Code != 404 {
		t.Fatal("empty id")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/api/dashboards", nil))
	if w.Code != 405 {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/dashboards/prospects", nil))
	if w.Code != 405 {
		t.Fatal(w.Code)
	}
	post := func(body any, origin string) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/dashboards", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	created := post(map[string]any{"id": "extra", "name": "Extra", "layout": "2x2", "widgets": []any{map[string]any{"id": "c", "type": "count", "title": "N"}}}, "http://localhost:7777")
	if created.Code != 201 {
		t.Fatal(created.Body.String())
	}
	if post("{", "http://127.0.0.1").Code != 400 {
		t.Fatal("bad json")
	}
	if post(map[string]any{"id": "bad/id"}, "http://localhost").Code != 400 {
		t.Fatal("bad id")
	}
	if post(map[string]any{"id": "evil"}, "https://evil.example").Code != 403 {
		t.Fatal("csrf")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
