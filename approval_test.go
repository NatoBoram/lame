package main_test

import (
	"testing"

	. "github.com/NatoBoram/lame"
)

func TestUnmarshalApproval(t *testing.T) {
	data := []byte(`{"explanation":"foo"}`)
	approval, err := UnmarshalApproval(data)
	if err != nil {
		t.Fatalf("couldn't unmarshal an approval: %s", err)
	}

	if approval.Explanation != "foo" {
		t.Fatalf("expected explanation to be 'foo', got %s", approval.Explanation)
	}
}

func TestMarshalApproval(t *testing.T) {
	approval := Approval{Explanation: "foo"}
	data, err := approval.Marshal()
	if err != nil {
		t.Fatalf("couldn't marshal an approval: %s", err)
	}

	expected := []byte(`{"explanation":"foo"}`)
	if string(data) != string(expected) {
		t.Fatalf("expected %s, got %s", expected, data)
	}
}
