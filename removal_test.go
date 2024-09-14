package main_test

import (
	"testing"

	. "github.com/NatoBoram/lame"
)

func TestUnmarshalRemoval(t *testing.T) {
	data := []byte(`{"reason":"foo"}`)
	removal, err := UnmarshalRemoval(data)
	if err != nil {
		t.Fatalf("couldn't unmarshal a removal: %s", err)
	}

	if removal.Reason != "foo" {
		t.Fatalf("expected reason to be 'foo', got %s", removal.Reason)
	}
}

func TestMarshalRemoval(t *testing.T) {
	removal := Removal{Reason: "foo"}
	data, err := removal.Marshal()
	if err != nil {
		t.Fatalf("couldn't marshal a removal: %s", err)
	}

	expected := []byte(`{"reason":"foo"}`)
	if string(data) != string(expected) {
		t.Fatalf("expected %s, got %s", expected, data)
	}
}
