package main_test

import (
	"testing"

	. "github.com/NatoBoram/lame"
)

func TestUnmarshalApproval(t *testing.T) {
	data := []byte(`{"someone":"foo","something":"bar","consequences":"baz","explanation":"qux"}`)
	approval, err := UnmarshalApproval(data)
	if err != nil {
		t.Fatalf("couldn't unmarshal an approval: %s", err)
	}

	if approval.Someone != "foo" {
		t.Fatalf("expected someone to be %s, got %s", "foo", approval.Someone)
	}

	if approval.Something != "bar" {
		t.Fatalf("expected something to be %s, got %s", "bar", approval.Something)
	}

	if approval.Consequences != "baz" {
		t.Fatalf("expected consequences to be %s, got %s", "baz", approval.Consequences)
	}

	if approval.Explanation != "qux" {
		t.Fatalf("expected explanation to be %s, got %s", "qux", approval.Explanation)
	}
}

func TestMarshalApproval(t *testing.T) {
	approval := Approval{
		Someone:      "foo",
		Something:    "bar",
		Consequences: "baz",
		Explanation:  "qux",
	}
	data, err := approval.Marshal()
	if err != nil {
		t.Fatalf("couldn't marshal an approval: %s", err)
	}

	expected := []byte(`{"someone":"foo","something":"bar","consequences":"baz","explanation":"qux"}`)
	if string(data) != string(expected) {
		t.Fatalf("expected %s, got %s", expected, data)
	}
}
