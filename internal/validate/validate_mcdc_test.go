package validate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/buger/recs/internal/record"
	"github.com/buger/recs/internal/validate"
)

// Verifies: SYS-REQ-260820-YWV4 SW-REQ-260820-8PMR
func TestCheckNilFieldAndNonEnum(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "crm.yaml"), []byte(`types:
  grant:
    required: [title]
    fields:
      status:
        enum: [open, closed]
      notes:
        type: string
`), 0o644); err != nil {
		t.Fatal(err)
	}
	rec := &record.Record{ID: "g1", Type: "grant", Fields: map[string]any{"title": "T", "notes": "x"}}
	res, err := validate.Check(root, []*record.Record{rec})
	if err != nil || !res.OK {
		t.Fatalf("%+v %v", res, err)
	}
	missing := &record.Record{ID: "g2", Type: "grant", Fields: map[string]any{"title": "T"}}
	res, err = validate.Check(root, []*record.Record{missing})
	if err != nil || !res.OK {
		t.Fatalf("missing optional %+v %v", res, err)
	}
}
