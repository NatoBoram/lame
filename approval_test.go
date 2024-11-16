package main_test

import (
	"testing"

	. "github.com/NatoBoram/lame"
)

func TestUnmarshalApproval(t *testing.T) {
	data := []byte(`{"someone":"foo","something":"bar","consequences":"baz"}`)
	approval, err := UnmarshalApproval(data)
	if err != nil {
		t.Fatalf("couldn't unmarshal an approval: %s", err)
	}

	if approval.Someone != "foo" {
		t.Errorf("expected someone to be %s, got %s", "foo", approval.Someone)
	}

	if approval.Something != "bar" {
		t.Errorf("expected something to be %s, got %s", "bar", approval.Something)
	}

	if approval.Consequences != "baz" {
		t.Errorf("expected consequences to be %s, got %s", "baz", approval.Consequences)
	}
}

func TestMarshalApproval(t *testing.T) {
	approval := Approval{
		Someone:      "foo",
		Something:    "bar",
		Consequences: "baz",
	}
	data, err := approval.Marshal()
	if err != nil {
		t.Fatalf("couldn't marshal an approval: %s", err)
	}

	expected := []byte(`{"someone":"foo","something":"bar","consequences":"baz"}`)
	if string(data) != string(expected) {
		t.Errorf("expected %s, got %s", expected, data)
	}
}
