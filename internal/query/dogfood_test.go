package query_test

import (
	"strings"
	"testing"

	"github.com/buger/recs/internal/query"
)

// Verifies: SW-REQ-260820-6EVX
// SW-REQ-260820-6EVX:malformed_input:negative
func TestParseRejectsDoubleEquals(t *testing.T) {
	_, err := query.Parse("type==grant")
	if err == nil || !strings.Contains(err.Error(), "unknown operator") {
		t.Fatalf("got %v", err)
	}
}
