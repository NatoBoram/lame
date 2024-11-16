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
