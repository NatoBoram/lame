package main

import (
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var approve = openai.FunctionDefinition{
	Name: "approve",
	Description: `Approve a post when the explanatory comment properly explains how someone is suffering consequences from something they voted for, supported or wanted to impose on other people.

The parameters of this function are goins to fill in the following template:

> <someone> voted for, supported or wanted to impose <something> on other people.
> <something> has the consequences of <consequences>.
> As a consequence of <something>, <consequences> happened to <someone>.

Only approve the post if filling this template would result in a coherent and plausible explanation. Otherwise, remove it.`,
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
				Description: "The consequences of the thing that the person voted for, supported or wanted to impose on other people and that they're suffering from. If the consequences haven't happened yet, remove the post.",
			},
		},
		Required: []string{"someone", "something", "consequences"},
	},
}

var remove = openai.FunctionDefinition{
	Name:        "remove",
	Description: "Remove a post when it corresponds to a removal reason",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"reason": {
				Type: jsonschema.String,
				Description: `These are the removal reasons of the subreddit. If the post fits one of these, remove it.

* The words "leopard", "ate", "face" and all their derivatives are forbidden in the title and explanatory comment.
* "Bad explanatory comment" means it is impossible to reconcile the explanatory comment with the required information.
* "No explanatory comment" means there's literally nothing in the explanatory comment. Do not approve posts without an explanatory comment.
* "No consequences yet" is for when the stated consequences didn't actually happen yet.`,
				Enum: []string{
					// Removal reasons
					string(ACTUAL_ANIMAL_ATTACK),
					string(BAD_EXPLANATORY_COMMENT),
					string(DIRECT_LINK_TO_OTHER_SUBREDDIT),
					string(DOES_NOT_FIT_THE_SUBREDDIT),
					string(LEOPARD_IN_TITLE_OR_EXPLANATORY_COMMENT),
					string(NO_CONSEQUENCES_YET),
					string(NO_EXPLANATORY_COMMENT),
					string(UNCIVIL_BEHAVIOUR),

					// Trapped flairs
					string(BYE_BYE_JOB),
					string(HYPOCRISY),
					string(LESSER_OF_TWO_EVILS),
					string(SELF_AWARE_WOLF),
					string(STUPIDITY),
					string(SUDDEN_BETRAYAL),
				},
			},
		},
		Required: []string{"reason"},
	},
}

var modFunctions = []openai.FunctionDefinition{remove, approve}
var modTools = []openai.Tool{
	{Type: openai.ToolTypeFunction, Function: &remove},
	{Type: openai.ToolTypeFunction, Function: &approve},
}
