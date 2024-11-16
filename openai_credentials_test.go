package main

import (
	"testing"
)

func TestUnmarshalOpenAiCredentials(t *testing.T) {
	data := []byte(`{"Token":"833a61a5-9493-46a5-b2fa-140e5736c3bb","BaseURL":"https://api.openai.com/v1","Model":"gpt-3.5-turbo"}`)
	result, err := UnmarshalOpenAiCredentials(data)
	if err != nil {
		t.Fatalf("Couldn't unmarshal OpenAI credentials: %v", err)
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
