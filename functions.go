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
				Type:        jsonschema.String,
				Description: "These are the rules of the subreddit. If the post violates one of these rules, remove it.",
				Enum: []string{
					"actual_animal_attack",
					"bad_explanatory_comment",
					"direct_link_to_other_subreddit",
					"does_not_fit_the_subreddit",
					"leopard_in_title_or_explanatory_comment",
					"no_explanatory_comment",
					"uncivil_behaviour",
				},
			},
		},
	},
}, {
	Name:        "approve",
	Description: "Approve a post when the explanatory comment explains how someone is suffering consequences from something they voted for, supported or wanted to impose on other people",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"explanation": {
				Type: jsonschema.String,
				Description: `Fill in the tags in the following sentences and make sure they are consistent:

> <Someone> voted for, supported or wanted to impose <something> on other people. <Something> has the consequences of <consequences>. As a consequence of <something>, <consequences> happened to <someone>.

Do not deviate from this template.`,
			},
		},
		Required: []string{"explanation"},
	},
}}
