package main

import (
	"testing"
)

func TestUnmarshalOpenAiCredentials(t *testing.T) {
	data := []byte(`{"Token":"833a61a5-9493-46a5-b2fa-140e5736c3bb","BaseURL":"https://api.openai.com/v1","Model":"gpt-3.5-turbo"}`)
	result, err := UnmarshalOpenAiCredentials(data)
	if err != nil {
		t.Fatalf("couldn't unmarshal OpenAI credentials: %v", err)
	}

	expected := OpenAiCredentials{
		Token:   "833a61a5-9493-46a5-b2fa-140e5736c3bb",
		BaseURL: "https://api.openai.com/v1",
		Model:   "gpt-3.5-turbo",
	}
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestMarshalOpenAiCredentials(t *testing.T) {
	creds := OpenAiCredentials{
		Token:   "70652f4a-61d2-4d78-bd3b-d4ee1c0ff296",
		BaseURL: "https://api.openai.com/v1",
		Model:   "gpt-3.5-turbo",
	}

	result, err := creds.Marshal()
	if err != nil {
		t.Fatalf("couldn't marshal OpenAI credentials: %v", err)
	}

	expected := `{"Token":"70652f4a-61d2-4d78-bd3b-d4ee1c0ff296","BaseURL":"https://api.openai.com/v1","Model":"gpt-3.5-turbo"}`
	if string(result) != expected {
		t.Errorf("expected %s, got %s", expected, string(result))
	}
}

func TestOpenAiCredentialsPath(t *testing.T) {
	configDir := "/tmp/config/lame"
	result := OpenAiCredentialsPath(configDir)

	expected := "/tmp/config/lame/openai_credentials.json"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestVerifyOpenAiCredentials(t *testing.T) {
	validCreds := OpenAiCredentials{
		Token:   "f8f920c4-4167-4d08-9d43-c686bad907c5",
		BaseURL: "https://api.openai.com/v1",
		Model:   "gpt-3.5-turbo",
	}

	if err := VerifyOpenAiCredentials(validCreds); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	invalidCreds := []OpenAiCredentials{
		{BaseURL: "https://api.openai.com/v1", Model: "gpt-3.5-turbo"},
		{BaseURL: "https://api.openai.com/v1", Token: "valid-token"},
		{Model: "gpt-3.5-turbo", Token: "valid-token"},
	}

	for _, creds := range invalidCreds {
		if err := VerifyOpenAiCredentials(creds); err == nil {
			t.Errorf("expected error, got nil for creds: %v", creds)
		}
	}
}
