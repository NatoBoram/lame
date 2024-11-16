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

func TestMarshalRedditCredentials(t *testing.T) {
	creds := RedditCredentials{
		ID:       "test-id",
		Secret:   "test-secret",
		Username: "test-username",
		Password: "test-password",
		Guide:    "test-guide",
	}

	result, err := creds.Marshal()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `{"ID":"test-id","Secret":"test-secret","Username":"test-username","Password":"test-password","Guide":"test-guide"}`
	if string(result) != expected {
		t.Errorf("expected %s, got %s", expected, string(result))
	}
}

func TestVerifyRedditCredentials(t *testing.T) {
	validCreds := RedditCredentials{
		ID:       "test-id",
		Secret:   "test-secret",
		Username: "test-username",
		Password: "test-password",
		Guide:    "test-guide",
	}

	if err := VerifyRedditCredentials(validCreds); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	invalidCreds := []RedditCredentials{
		{Secret: "test-secret", Username: "test-username", Password: "test-password", Guide: "test-guide"},
		{ID: "test-id", Username: "test-username", Password: "test-password", Guide: "test-guide"},
		{ID: "test-id", Secret: "test-secret", Password: "test-password", Guide: "test-guide"},
		{ID: "test-id", Secret: "test-secret", Username: "test-username", Guide: "test-guide"},
		{ID: "test-id", Secret: "test-secret", Username: "test-username", Password: "test-password"},
	}

	for _, creds := range invalidCreds {
		if err := VerifyRedditCredentials(creds); err == nil {
			t.Errorf("expected error, got nil for creds: %v", creds)
		}
	}
}
