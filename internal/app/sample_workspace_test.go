package app_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/buger/recs/internal/app"
)

func sampleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "testdata", "sample"))
	if _, err := os.Stat(filepath.Join(root, "crm.yaml")); err != nil {
		t.Fatal(root, err)
	}
	return root
}

// Verifies: SYS-REQ-260820-456X SW-REQ-260820-NA06 INT-REQ-260820-NHBY
func TestSampleWorkspaceDashboards(t *testing.T) {
	a, err := app.Open(sampleRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	recs, err := a.List("")
	if err != nil || len(recs) < 10 {
		t.Fatalf("records: %v count=%d", err, len(recs))
	}
	dashboards, err := a.ListDashboards()
	if err != nil || len(dashboards) < 4 {
		t.Fatalf("dashboards: %v count=%d", err, len(dashboards))
	}
	want := map[string]bool{"prospects": false, "workspace": false, "inbox": false, "grants": false}
	for _, d := range dashboards {
		if _, ok := want[d.ID]; ok {
			want[d.ID] = true
		}
		p, err := a.ProjectDashboard(d.ID)
		if err != nil {
			t.Fatalf("project %s: %v", d.ID, err)
		}
		if p.WidgetCount == 0 {
			t.Fatalf("project %s: empty widgets", d.ID)
		}
	}
	for id, seen := range want {
		if !seen {
			t.Fatalf("missing dashboard %s", id)
		}
	}
}
