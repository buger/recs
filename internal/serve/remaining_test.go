package serve_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"crm/internal/serve"
)

// Verifies: SYS-REQ-260821-QF1J SW-REQ-260821-82BA INT-REQ-260821-MRGW STK-REQ-260821-NTWY
// SYS-REQ-260821-QF1J:nominal:nominal
// SW-REQ-260821-82BA:nominal:nominal
// INT-REQ-260821-MRGW:nominal:nominal
// INT-REQ-260821-MRGW:integration:integration
// STK-REQ-260821-NTWY:nominal:nominal
// MCDC SYS-REQ-260821-QF1J: attachment_served=F, board_filters_applied=F, listen_port_GT_0=T, record_editor_saved=F, record_view_rendered=F, remaining_ui_requested=T, search_page_shown=F, ui_request_rejected=F, wikilink_resolved=T => TRUE
// MCDC SW-REQ-260821-82BA: attachment_served=F, board_filters_applied=F, listen_port_GE_1=T, listen_port_LE_65535=T, record_editor_saved=F, record_view_rendered=F, remaining_ui_requested=T, search_page_shown=F, ui_request_rejected=F, wikilink_resolved=T => TRUE
// MCDC INT-REQ-260821-MRGW: attachment_served=F, board_filters_applied=F, http_api_requested=T, record_editor_saved=F, record_view_rendered=F, remaining_ui_requested=T, search_page_shown=F, shared_app_layer_used=T, shared_record_model_used=T, ui_request_rejected=F, wikilink_resolved=T => TRUE
//mcdc:ignore SYS-REQ-260821-QF1J: attachment_served=F, board_filters_applied=F, listen_port_GT_0=T, record_editor_saved=F, record_view_rendered=F, remaining_ui_requested=T, search_page_shown=F, ui_request_rejected=F, wikilink_resolved=F => FALSE -- remaining UI without a rendered or rejected outcome is the literal negation of the remaining-UI contract [reviewed: agent:grok] [category: defensive]
//mcdc:ignore SW-REQ-260821-82BA: attachment_served=F, board_filters_applied=F, listen_port_GE_1=T, listen_port_LE_65535=T, record_editor_saved=F, record_view_rendered=F, remaining_ui_requested=T, search_page_shown=F, ui_request_rejected=F, wikilink_resolved=F => FALSE -- remaining UI without a rendered or rejected outcome is the literal negation of the remaining-UI contract [reviewed: agent:grok] [category: defensive]
//mcdc:ignore INT-REQ-260821-MRGW: attachment_served=F, board_filters_applied=F, http_api_requested=T, record_editor_saved=F, record_view_rendered=F, remaining_ui_requested=T, search_page_shown=F, shared_app_layer_used=T, shared_record_model_used=T, ui_request_rejected=F, wikilink_resolved=F => FALSE -- remaining UI on the shared layer without a rendered outcome is the literal negation of the integration contract [reviewed: agent:grok] [category: defensive]
func TestRecordViewEditorSearchFiltersAttachments(t *testing.T) {
	a := setupApp(t)
	if _, err := a.Create("person", "person_alice", map[string]any{"name": "Alice Smith"}, "Hello [[grant_a]]\n"); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(a.Root(), "attachments", "doc.txt"), []byte("file"), 0o644)
	if _, err := a.Edit("grant_a", map[string]any{"attachments": []any{"attachments/doc.txt"}, "owner": "leonid"}, nil, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Link("person_alice", "grant_a", "follows"); err != nil {
		t.Fatal(err)
	}
	h := serve.Handler(a)
	get := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w
	}
	ui := get("/")
	if ui.Code != 200 || !strings.Contains(ui.Body.String(), "#/r/") || !strings.Contains(ui.Body.String(), "#/search") || !strings.Contains(ui.Body.String(), "nav-search") {
		t.Fatal("ui routes")
	}
	rec := get("/api/records/person_alice")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "html") || !strings.Contains(rec.Body.String(), "relations") {
		t.Fatal(rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "grant_a") {
		t.Fatal("wikilink/relation", rec.Body.String())
	}
	board := get("/api/boards/grants?owner=leonid")
	if board.Code != 200 || !strings.Contains(board.Body.String(), "filter_controls") {
		t.Fatal(board.Body.String())
	}
	file := get("/api/files/attachments/doc.txt")
	if file.Code != 200 || file.Body.String() != "file" {
		t.Fatal(file.Code, file.Body.String())
	}
	if get("/api/files/../crm.yaml").Code == 200 && strings.Contains(get("/api/files/../crm.yaml").Body.String(), "name:") {
		t.Fatal("traversal")
	}
	if get("/api/files/nope").Code != 404 {
		t.Fatal("missing file")
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/records/grant_a", bytes.NewReader([]byte(`{"body":"patched body","status":"preparing"}`)))
	req.Header.Set("Origin", "http://localhost")
	h.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	got, _ := a.Show("grant_a")
	if strings.TrimSpace(got.Body) != "patched body" {
		t.Fatal(got.Body)
	}
	search := get("/api/search?q=Alice")
	if search.Code != 200 || !strings.Contains(search.Body.String(), "person_alice") {
		t.Fatal(search.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/files/attachments/doc.txt", nil))
	if w.Code != 405 {
		t.Fatal(w.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil || payload["ok"] != true {
		t.Fatal(payload)
	}
}
