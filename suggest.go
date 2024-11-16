package main

import (
	"fmt"

	"github.com/logrusorgru/aurora/v4"
	openai "github.com/sashabaranov/go-openai"
)

// Exported ToolName function
func ToolName(resp openai.ChatCompletionResponse) *string {
	if len(resp.Choices) == 0 {
		fmt.Println("No suggestions.")
		prettyPrint(resp)
		return nil
	}

	if resp.Choices[0].Message.ToolCalls == nil {
		fmt.Println("No tools were called.")
		prettyPrint(resp)
		return nil
	}

	if len(resp.Choices) > 1 {
		fmt.Println("More than one choice.")
		prettyPrint(resp)
	}

	if len(resp.Choices[0].Message.ToolCalls) > 1 {
		fmt.Println("More than one tool call.")
		prettyPrint(resp.Choices[0].Message.ToolCalls)
	}

	return &resp.Choices[0].Message.ToolCalls[0].Function.Name
}

func toolCall(resp openai.ChatCompletionResponse) (*Approval, *Removal, error) {
	name := ToolName(resp)
	if name == nil {
		return nil, nil, nil
	}

	function := resp.Choices[0].Message.ToolCalls[0].Function
	switch function.Name {

	case "approve":
		approval, err := UnmarshalApproval([]byte(function.Arguments))
		if err != nil {
			return &approval, nil, fmt.Errorf("failed to unmarshal approval: %w", err)
		}
		return &approval, nil, nil

	case "remove":
		removal, err := UnmarshalRemoval([]byte(function.Arguments))
		if err != nil {
			return nil, &removal, fmt.Errorf("failed to unmarshal removal: %w", err)
		}
		return nil, &removal, nil
	}

	return nil, nil, fmt.Errorf("unknown tool: %s", function.Name)
}

func suggest(resp openai.ChatCompletionResponse) (*Approval, *Removal, error) {
	approval, removal, error := toolCall(resp)
	if error != nil {
		return approval, removal, fmt.Errorf("failed to get tool call: %w", error)
	}

	if approval != nil {
		suggestApprove(*approval)
	}

	if removal != nil {
		suggestRemove(*removal)
	}

	return approval, removal, nil
}

func suggestApprove(approval Approval) {
	explanation := fmt.Sprintf(`1. %s voted for, supported or wanted to impose %s on other people.
2. %s has the consequences of %s.
3. As a consequence of %s, %s happened to %s.`,
		aurora.Bold(approval.Someone), aurora.Bold(approval.Something),
		aurora.Bold(approval.Something), aurora.Bold(approval.Consequences),
		aurora.Bold(approval.Something), aurora.Bold(approval.Consequences), aurora.Bold(approval.Someone),
	)

	fmt.Printf(`Recommendation: %s
Someone: %s
Something: %s
Consequences: %s

%s

`,
		aurora.Green("Approve"),
		aurora.Gray(6, approval.Someone),
		aurora.Gray(6, approval.Something),
		aurora.Gray(6, approval.Consequences),
		explanation,
	)
}

func suggestRemove(removal Removal) {
	fmt.Printf(`Recommendation: %s
Reason: %s

`, aurora.Red("Remove"), removal.Reason)
}
