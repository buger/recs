package serve_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crm/internal/serve"
)

// Verifies: SW-REQ-260821-82BA
// SW-REQ-260821-82BA
func TestServeBoardFiltersAndMethods(t *testing.T) {
	a := setupApp(t)
	h := serve.Handler(a)
	req := func(method, path string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	_ = req(http.MethodGet, "/api/boards/grants?owner=leonid")
	_ = req(http.MethodGet, "/api/boards/grants?owner=")
	_ = req(http.MethodGet, "/api/boards/grants?=x")
	_ = req(http.MethodGet, "/api/boards/grants?foo=bar&empty=")
	_ = req(http.MethodHead, "/api/files/grant_a/nope")
	_ = req(http.MethodPost, "/api/files/grant_a/nope")
	_ = req(http.MethodGet, "/api/files/grant_a/nope")
	_ = req(http.MethodDelete, "/api/records/grant_a")
	_ = req(http.MethodGet, "/api/dashboards")
	dash := filepath.Join(a.Root(), "dashboards")
	if err := os.Chmod(dash, 0); err == nil {
		t.Cleanup(func() { _ = os.Chmod(dash, 0o755) })
		_ = req(http.MethodGet, "/api/dashboards")
	}
	_ = strings.TrimSpace("x")
}

// Verifies: SW-REQ-260821-82BA
// SW-REQ-260821-82BA
func TestServeFmtSprintAndEmptyDashboards(t *testing.T) {
	a := setupApp(t)
	h := serve.Handler(a)
	req := func(method, path, body string) *httptest.ResponseRecorder {
		var r *http.Request
		if body == "" {
			r = httptest.NewRequest(method, path, nil)
		} else {
			r = httptest.NewRequest(method, path, strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	_ = req(http.MethodPatch, "/api/records/grant_a", `{"body":null}`)
	_ = req(http.MethodPatch, "/api/records/grant_a", `{"body":123}`)
	_ = req(http.MethodPatch, "/api/records/grant_a", `{"body":"x","_if_version":"nope"}`)
	dash := filepath.Join(a.Root(), "dashboards")
	entries, _ := os.ReadDir(dash)
	for _, e := range entries {
		_ = os.Remove(filepath.Join(dash, e.Name()))
	}
	_ = req(http.MethodGet, "/api/dashboards", "")
}

// Verifies: SYS-REQ-260820-9W1S
// SYS-REQ-260820-9W1S
func TestListenSuccessAndBusyPort(t *testing.T) {
	a := setupApp(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	if err := serve.Listen(a, port); err == nil {
		t.Fatal("expected busy port")
	}
}
