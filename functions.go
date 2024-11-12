package main

import (
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var modFunctions = []openai.FunctionDefinition{{
	Name:        "remove",
	Description: "Remove a post when it violates a rule",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"reason": {
				Type: jsonschema.String,
				Description: `These are the rules of the subreddit. If the post violates one of these rules, remove it.

* The words "leopard", "ate", "face" and all their derivatives are forbidden in the title and explanatory comment.
* "Bad explanatory comment" means it is impossible to reconcile the explanatory comment with the required information. Namely, who's suffering from what consequences and what did they support?
* "No consequences yet" is for when the stated consequences didn't actually happen yet.`,
				Enum: []string{
					string(ACTUAL_ANIMAL_ATTACK),
					string(BAD_EXPLANATORY_COMMENT),
					string(DIRECT_LINK_TO_OTHER_SUBREDDIT),
					string(DOES_NOT_FIT_THE_SUBREDDIT),
					string(LEOPARD_IN_TITLE_OR_EXPLANATORY_COMMENT),
					string(NO_CONSEQUENCES_YET),
					string(NO_EXPLANATORY_COMMENT),
					string(UNCIVIL_BEHAVIOUR),
				},
			},
		},
	},
}, {
	Name:        "approve",
	Description: "Approve a post when the explanatory comment explains how someone is suffering consequences from something they voted for, supported or wanted to impose on other people.",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"someone": {
				Type:        jsonschema.String,
				Description: "The name of the person who voted for, supported or wanted to impose something on other people and who's suffering consequences of it.",
			},
			"something": {
				Type:        jsonschema.String,
				Description: "The thing that the person voted for, supported or wanted to impose on other people.",
			},
			"consequences": {
				Type:        jsonschema.String,
				Description: "The consequences of the thing that the person voted for, supported or wanted to impose on other people and that they're suffering from.",
			},
		},
		Required: []string{"someone", "something", "consequences"},
	},
}}
