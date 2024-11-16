package main

import (
	"testing"
)

func TestFlairToRemovalReason(t *testing.T) {
	tests := []struct {
		input    RemovalReason
		expected RemovalReason
	}{
		{ACTUAL_ANIMAL_ATTACK, ACTUAL_ANIMAL_ATTACK},
		{SELF_AWARE_WOLF, DOES_NOT_FIT_THE_SUBREDDIT},
	}

	for _, test := range tests {
		result := FlairToRemovalReason(test.input)
		if result != test.expected {
			t.Errorf("FlairToRemovalReason(%s) = %s; expected %s", test.input, result, test.expected)
		}
	}
}

func TestFormatRemovalMessage(t *testing.T) {
	removalReason := NO_EXPLANATORY_COMMENT
	model := "gpt-3.5-turbo"

	result, err := FormatRemovalMessage(removalReason, model)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := `Thank you for your submission! Unfortunately, it has been removed for the following reason:

* **Rule 3 :** Write an [explanatory comment](https://www.reddit.com/r/LeopardsAteMyFace/comments/lt8zlq)

*This removal was LLM-assisted. See the source code at <https://github.com/NatoBoram/lame>. Model: ` + "`gpt-3.5-turbo`" + `.*

*If you have any questions or concerns about this removal, please feel free to [message the moderators](https://www.reddit.com/message/compose/?to=/r/LeopardsAteMyFace) thru Modmail. Thanks!*`
	if result != expected {
		t.Errorf("FormatRemovalMessage(%v, %s) = %s; expected %s", removalReason, model, result, expected)
	}
}
