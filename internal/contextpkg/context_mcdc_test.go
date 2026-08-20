package contextpkg

import (
	"testing"

	"crm/internal/record"
)

func TestAssembleSeenAndLooksIDFalse(t *testing.T) {
	seed := &record.Record{ID: "person_a", Fields: map[string]any{
		"company":  "plain",
		"people":   []any{"notanid"},
		"related":  []any{"alsoplain"},
		"relations": []any{map[string]any{"target": "plainname"}, map[string]any{"target": "person_a"}},
	}}
	got := Assemble(seed, []*record.Record{seed})
	if len(got.Related) != 0 {
		t.Fatalf("%#v", got.Related)
	}
	_ = outgoing(&record.Record{Fields: map[string]any{"company": "x", "people": []any{"y"}, "relations": []any{map[string]any{"target": "z"}}}})
}
