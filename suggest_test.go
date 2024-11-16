package main

import (
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestToolName(t *testing.T) {
	resp := openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{
		Message: openai.ChatCompletionMessage{ToolCalls: []openai.ToolCall{{
			Function: openai.FunctionCall{Name: "approve"},
		}}},
	}}}

	result := ToolName(resp)
	if result == nil || *result != "approve" {
		t.Errorf("ToolName(%v) = %s; expected %s", resp, *result, "approve")
	}
}
