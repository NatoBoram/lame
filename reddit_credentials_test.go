package main

import (
	"testing"
)

func TestUnmarshalRedditCredentials(t *testing.T) {
	data := []byte(`{"ID":"test-id","Secret":"test-secret","Username":"test-username","Password":"test-password","Guide":"test-guide"}`)
	result, err := UnmarshalRedditCredentials(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := RedditCredentials{
		ID:       "test-id",
		Secret:   "test-secret",
		Username: "test-username",
		Password: "test-password",
		Guide:    "test-guide",
	}
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
