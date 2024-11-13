package main

import (
	"fmt"

	"github.com/logrusorgru/aurora/v4"
	openai "github.com/sashabaranov/go-openai"
)

func suggest(resp openai.ChatCompletionResponse) (*Approval, *Removal, error) {
	if len(resp.Choices) == 0 {
		fmt.Println("No suggestions.")
		prettyPrint(resp)
		return nil, nil, nil
	}

	if resp.Choices[0].Message.ToolCalls == nil {
		fmt.Println("No tools were called.")
		prettyPrint(resp)
		return nil, nil, nil
	}

	function := resp.Choices[0].Message.ToolCalls[0].Function
	switch function.Name {

	case "approve":
		approval, err := UnmarshalApproval([]byte(function.Arguments))
		if err != nil {
			return &approval, nil, fmt.Errorf("failed to unmarshal approval: %w", err)
		}

		suggestApprove(approval)
		return &approval, nil, nil

	case "remove":
		removal, err := UnmarshalRemoval([]byte(function.Arguments))
		if err != nil {
			return nil, &removal, fmt.Errorf("failed to unmarshal removal: %w", err)
		}

		suggestRemove(removal)
		return nil, &removal, nil
	}

	return nil, nil, fmt.Errorf("unknown function: %s", function.Name)
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
