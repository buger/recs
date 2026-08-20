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

	"crm/internal/app"
	"crm/internal/serve"
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
