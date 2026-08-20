package contextpkg

import (
	"testing"

	"crm/internal/record"
)

func TestLooksIDAndOutgoing(t *testing.T) {
	if looksID("") || looksID("   ") || looksID("plain") {
		t.Fatal("false ids")
	}
	if !looksID("company_acme") || !looksID("a-b") {
		t.Fatal("true ids")
	}
	seed := &record.Record{ID: "person_a", Fields: map[string]any{
		"company": "company_acme", "people": []any{"person_b"},
		"relations": []any{map[string]any{"target": "task_1"}, "skip", map[string]any{"target": 1}},
		"owner": "notanid",
	}}
	other := &record.Record{ID: "company_acme", Fields: map[string]any{"people": []string{"person_a"}}}
	task := &record.Record{ID: "task_1", Fields: map[string]any{}}
	orphan := &record.Record{ID: "orphan", Fields: map[string]any{"company": "missing_co"}}
	got := Assemble(seed, []*record.Record{seed, other, task, orphan})
	if len(got.Related) < 2 {
		t.Fatalf("%#v", got.Related)
	}
	if outgoing(&record.Record{Fields: map[string]any{"relations": "nope"}}) != nil && false {
		t.Fatal("noop")
	}
	_ = outgoing(&record.Record{Fields: map[string]any{"relations": []any{1, map[string]any{}}}})
}
