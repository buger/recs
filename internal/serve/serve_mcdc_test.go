package serve_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"crm/internal/app"
	"crm/internal/serve"
)

// Verifies: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7 SYS-REQ-260821-QF1J SW-REQ-260821-82BA
func TestHandlerErrorAndMethodIndependence(t *testing.T) {
	a := setupApp(t)
	h := serve.Handler(a)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodHead, "/", nil))
	if w.Code != 200 {
		t.Fatal(w.Code)
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/boards/grants/move", nil))
	if w.Code != 405 && w.Code != 404 && w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/records/grant_a", bytes.NewReader([]byte(`{"status":"nope"}`)))
	req.Header.Set("Origin", "http://localhost")
	h.ServeHTTP(w, req)
	if w.Code == 200 {
		t.Fatal("invalid enum should fail")
	}

	records := filepath.Join(a.Root(), "records")
	if err := os.Chmod(records, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(records, 0o755) })
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/records", nil))
	if w.Code != 500 {
		t.Fatal("list err", w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/search?q=A", nil))
	if w.Code != 500 {
		t.Fatal("search err", w.Code)
	}

	boards := filepath.Join(a.Root(), "boards")
	_ = os.Chmod(records, 0o755)
	if err := os.RemoveAll(boards); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(boards, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/boards", nil))
	if w.Code != 500 {
		t.Fatal("boards err", w.Code, w.Body.String())
	}
}

// Verifies: SYS-REQ-260820-9W1S SW-REQ-260820-8ZS7 INT-REQ-260820-AHKR
func TestListenSuccess(t *testing.T) {
	a := app.OpenOrCWD(t.TempDir())
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	errc := make(chan error, 1)
	go func() { errc <- serve.Listen(a, 0) }()
	select {
	case err := <-errc:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(150 * time.Millisecond):
		// listen succeeded and Serve is blocking — that is the missing false branch
	}
}
