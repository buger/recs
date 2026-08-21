package record

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

type strer struct{ s string }

func (s strer) String() string { return s.s }

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func TestParseVariants(t *testing.T) {
	rec, err := Parse("x.md", []byte("no frontmatter"))
	if err != nil || rec.Body != "no frontmatter" {
		t.Fatalf("%v %#v", err, rec)
	}
	rec, err = Parse("x.md", []byte("---\r\nid: a\r\ntype: note\r\n---\r\nbody"))
	if err != nil || rec.ID != "a" || rec.Type != "note" {
		t.Fatalf("%v %#v", err, rec)
	}
	rec, err = Parse("x.md", []byte("---\nid: b\ntype: note\n---\n"))
	if err != nil || rec.ID != "b" {
		t.Fatal(err)
	}
	if _, err = Parse("x.md", []byte("---\nid: z\n")); err == nil {
		t.Fatal("expected missing closer")
	}
	if _, err = Parse("x.md", []byte("---\n:\n---\n")); err == nil {
		t.Fatal("expected bad yaml")
	}
	rec, err = Parse("x.md", []byte("---\n\n---\n"))
	if err != nil || rec.Fields == nil {
		t.Fatal(err)
	}
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func TestBytesAndGetSetDelete(t *testing.T) {
	r := &Record{}
	_ = r.Bytes()
	r = &Record{ID: "n_1", Type: "note", Fields: nil, Body: "hi"}
	raw := string(r.Bytes())
	if !strings.Contains(raw, "id: n_1") || !strings.HasSuffix(raw, "\n") {
		t.Fatal(raw)
	}
	if r.Get("missing") != nil || r.GetString("missing") != "" {
		t.Fatal("empty")
	}
	r.Set("title", "Hello")
	if r.GetString("title") != "Hello" {
		t.Fatal(r.Get("title"))
	}
	r.Set("id", "n_2")
	r.Set("type", "task")
	if r.ID != "n_2" || r.Type != "task" {
		t.Fatal(r.ID, r.Type)
	}
	r.Set("id", 1)
	r.Set("type", 2)
	r.Set("amount.requested", 10)
	if r.Get("amount.requested") != 10 {
		t.Fatal(r.Get("amount.requested"))
	}
	if r.Get("amount.missing") != nil || r.Get("title.nested") != nil {
		t.Fatal("nested miss")
	}
	r.Delete("nope")
	r.Delete("title")
	r.Delete("amount.requested")
	r.Delete("amount.missing.x")
	empty := &Record{}
	empty.Delete("x")
	empty.Delete("a.b")
	if empty.GetString("x") != "" {
		t.Fatal("empty delete")
	}
	r.Fields["n"] = 3
	r.Fields["s"] = strer{"z"}
	if r.GetString("n") != "3" || r.GetString("s") != "z" {
		t.Fatal(r.GetString("n"), r.GetString("s"))
	}
	_ = r.Version()
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func TestAsMapAndValid(t *testing.T) {
	m, ok := asMap(map[string]any{"a": 1})
	if !ok || m["a"] != 1 {
		t.Fatal(m, ok)
	}
	m, ok = asMap(map[any]any{1: "x"})
	if !ok || m["1"] != "x" {
		t.Fatal(m, ok)
	}
	if _, ok = asMap("no"); ok {
		t.Fatal("string map")
	}
	if ValidStableID("") || ValidStableID("a/b") || ValidStableID("a..b") || ValidStableID("A") {
		t.Fatal("invalid id accepted")
	}
	if !ValidStableID("grant_x") {
		t.Fatal("valid id")
	}
	if ValidType("") || ValidType("a\\b") || ValidType("..") || !ValidType("grant") {
		t.Fatal("type")
	}
	if SlugID("", "") == "" || SlugID("Grant", "Hello World") != "grant_hello_world" {
		t.Fatal(SlugID("Grant", "Hello World"))
	}
}

// Verifies: SYS-REQ-260820-9J7C SW-REQ-260820-N02Y
func TestTypeDirDisplayCompareSlice(t *testing.T) {
	for typ, dir := range map[string]string{
		"person": "people", "company": "companies", "customer": "customers",
		"grant": "grants", "application": "applications", "onboarding": "onboarding",
		"email": "emails", "task": "tasks", "note": "notes", "": "misc", "foo": "foos",
	} {
		if TypeDir(typ) != dir {
			t.Fatalf("%s -> %s", typ, TypeDir(typ))
		}
	}
	r := &Record{ID: "x", Path: "/tmp/z.md", Fields: map[string]any{}}
	if DisplayName(r) != "x" {
		t.Fatal(DisplayName(r))
	}
	r.ID = ""
	if DisplayName(r) != "z.md" {
		t.Fatal(DisplayName(r))
	}
	r.Fields["name"] = "N"
	if DisplayName(r) != "N" {
		t.Fatal(DisplayName(r))
	}
	r.Fields["title"] = "T"
	if DisplayName(r) != "T" {
		t.Fatal(DisplayName(r))
	}
	if CompareValues(1, 2) >= 0 || CompareValues(2, 1) <= 0 || CompareValues(2, 2) != 0 {
		t.Fatal("nums")
	}
	if CompareValues(int64(1), float32(2)) >= 0 || CompareValues(uint64(3), float64(1)) <= 0 {
		t.Fatal("typed nums")
	}
	if CompareValues("2020-01-01", "2021-01-01") >= 0 || CompareValues("2021-01-02T00:00:00Z", "2021-01-01T00:00:00Z") <= 0 {
		t.Fatal("times")
	}
	if CompareValues("2020-01-01T00:00:00Z", "2020-01-01T00:00:00Z") != 0 {
		t.Fatal("eq time")
	}
	if CompareValues("a", "B") >= 0 {
		t.Fatal("str")
	}
	if CompareValues(nil, "x") == 999 {
		t.Fatal("nil")
	}
	_ = time.Now()
	if StringSlice(nil) != nil || len(StringSlice([]string{"a"})) != 1 {
		t.Fatal("slice")
	}
	if len(StringSlice([]any{1, 2})) != 2 || StringSlice("") != nil || StringSlice("x")[0] != "x" {
		t.Fatal("slice2")
	}
	if _, _, ok := asFloat(nil); ok {
		t.Fatal("nil float")
	}
	if _, _, ok := asFloat("nope"); ok {
		t.Fatal("bad float")
	}
	if _, ok := asTime("not-a-date"); ok {
		t.Fatal("bad time")
	}
	if fmt.Sprint(sanitizeID(" Hello!!World ")) != "hello_world" {
		t.Fatal(sanitizeID(" Hello!!World "))
	}
}
