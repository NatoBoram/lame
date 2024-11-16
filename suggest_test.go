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

func TestToolCall_Approve(t *testing.T) {
	resp := openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{
		Message: openai.ChatCompletionMessage{ToolCalls: []openai.ToolCall{{
			Function: openai.FunctionCall{
				Name:      "approve",
				Arguments: `{"someone":"someone","something":"something","consequences":"consequences"}`,
			},
		}}},
	}}}

	approval, removal, err := ToolCall(resp)
	if err != nil {
		t.Errorf("ToolCall(%v) returned error: %v", resp, err)
	}

	expected := Approval{
		Someone:      "someone",
		Something:    "something",
		Consequences: "consequences",
	}
	if approval == nil || *approval != expected || removal != nil {
		t.Errorf("ToolCall(%v) = %v, %v; expected %v", resp, approval, removal, expected)
	}
}

func TestToolCall_Remove(t *testing.T) {
	resp := openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{
		Message: openai.ChatCompletionMessage{ToolCalls: []openai.ToolCall{{
			Function: openai.FunctionCall{
				Name:      "remove",
				Arguments: `{"reason":"does_not_fit_the_subreddit"}`,
			},
		}}},
	}}}

	approval, removal, err := ToolCall(resp)
	if err != nil {
		t.Errorf("ToolCall(%v) returned error: %v", resp, err)
	}

	expected := Removal{Reason: DOES_NOT_FIT_THE_SUBREDDIT}
	if removal == nil || *removal != expected || approval != nil {
		t.Errorf("ToolCall(%v) = %v, %v; expected %v", resp, approval, removal, expected)
	}
}

func TestToolCall_Invalid(t *testing.T) {
	resp := openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{{
		Message: openai.ChatCompletionMessage{ToolCalls: []openai.ToolCall{{
			Function: openai.FunctionCall{Name: "invalid_tool", Arguments: `{}`},
		}}},
	}}}

	approval, removal, err := ToolCall(resp)
	if err == nil || err.Error() != "unknown tool: invalid_tool" {
		t.Errorf("ToolCall(%v) returned error: %v; expected unknown tool error", resp, err)
	}
	if approval != nil || removal != nil {
		t.Errorf("ToolCall(%v) = %v, %v; expected nil, nil", resp, approval, removal)
	}
}
