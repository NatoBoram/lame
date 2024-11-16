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
