package record

import (
	"testing"
)

func TestParseCRFirstCloserAndNullFields(t *testing.T) {
	rec, err := Parse("x.md", []byte("---\nid: a\n---\r\nmore\n---\nbody"))
	if err != nil || rec.ID != "a" {
		t.Fatalf("%v %#v", err, rec)
	}
	rec, err = Parse("x.md", []byte("---\nid: b\n---\nmore\n---\r\ntail"))
	if err != nil || rec.ID != "b" {
		t.Fatalf("%v %#v", err, rec)
	}
	rec, err = Parse("x.md", []byte("---\nnull\n---\n"))
	if err != nil || rec.Fields == nil {
		t.Fatalf("null fields %v %#v", err, rec)
	}
}

func TestSetNilFieldsAndExistingMap(t *testing.T) {
	r := &Record{}
	r.Set("title", "Hello")
	if r.GetString("title") != "Hello" || r.Fields == nil {
		t.Fatal(r)
	}
	r.Set("amount.requested", 10)
	r.Set("amount.currency", "USD")
	if r.Get("amount.requested") != 10 || r.Get("amount.currency") != "USD" {
		t.Fatal(r.Fields)
	}
}

func TestValidTypeSanitizeAndCompareIndependence(t *testing.T) {
	if ValidType("Grant") {
		t.Fatal("uppercase type should fail sanitize equality")
	}
	if sanitizeID("{hello}") != "hello" {
		t.Fatal(sanitizeID("{hello}"))
	}
	if sanitizeID("!hello") != "hello" {
		t.Fatal(sanitizeID("!hello"))
	}
	if sanitizeID("!!") != "" {
		t.Fatal(sanitizeID("!!"))
	}
	if CompareValues(1, "abc") == 999 {
		t.Fatal("num vs text")
	}
	if CompareValues("2020-01-01", "not-a-date") == 999 {
		t.Fatal("time vs text")
	}
	_ = CompareValues("plain", "2020-01-01")
}
