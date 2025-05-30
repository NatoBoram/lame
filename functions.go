package main

import (
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var approve = openai.FunctionDefinition{
	Name: "approve",
	Description: `Approve a post when the explanatory comment properly explains how someone is suffering consequences from something they voted for, supported or wanted to impose on other people and it does not correspond to any removal reasons.

The parameters of this function are goins to fill in the following template:

> <someone> voted for, supported or wanted to impose <something> on other people.
> <something> has the consequences of <consequences>.
> As a consequence of <something>, <consequences> happened to <someone>.

Only approve the post if filling this template would result in a coherent and plausible explanation. Otherwise, remove it.`,
	Strict: true,
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"someone": {
				Type:        jsonschema.String,
				Description: "The name of the person who voted for, supported or wanted to impose something on other people and who's suffering consequences of it.",
			},
			"something": {
				Type:        jsonschema.String,
				Description: "The thing that the person voted for, supported or wanted to impose on other people. Max 40 characters.",
			},
			"consequences": {
				Type:        jsonschema.String,
				Description: "The consequences of the thing that the person voted for, supported or wanted to impose on other people and that they're suffering from. If the consequences haven't happened yet, remove the post. Max 40 characters.",
			},
		},
		Required:             []string{"someone", "something", "consequences"},
		AdditionalProperties: false,
	},
}

var remove = openai.FunctionDefinition{
	Name:        "remove",
	Description: "Remove a post when it corresponds to a removal reason",
	Strict:      true,
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"reason": {
				Type: jsonschema.String,
				Description: `These are the removal reasons of the subreddit. If the post fits one of these, remove it.

bad_explanatory_comment: It is impossible to identify who supported something or what they supported or what are the consequences from the explanatory comment.
direct_link_to_other_subreddit: Contains a reference to another subreddit.
distinct_enabler_and_victim: The person who supported something is not the same person as the one who receives the consequences. For example, parents not vaccinating their child and then that child getting sick. The child is an innocent victim of what their parents imposed on them.
does_not_fit_the_subreddit: The post is not about someone who's suffering consequences from something they voted for or supported or wanted to impose on other people.
future_consequences: The consequences have not happened yet. They may be likely to maybe potentially happen in the future, but they didn't happen yet.
leopard_in_title_or_explanatory_comment: The words "leopards", "ate" and "face" and all their derivatives are forbidden in the title, body and explanatory comment. For example, no cats munching on visages either.
no_consequences: There are no consequences in the post or explanatory comment. For example, being shocked, feeling regrets and getting criticized aren't consequences.
no_explanatory_comment: The explanatory comment is empty.

bye_bye_job: Someone did something and lost their job as a consequence, but losing their job isn't a consequence of what they supported. For example, someone was fired for not getting vaccinated, but they didn't support firing people for not getting vaccinated.
hypocrisy: Someone is being a hypocrite but they're not receiving any consequences of what they supported.
lesser_of_two_evils: Someone voted for something terrible, but that's only because the other choice was something even worse.
self_aware_wolf: Someone accidentally describes themselves but they're not self-aware enough to realize it.
stupidity: Someone is being stupid, but there's no schadenfreude to be had.
sudden_betrayal: Someone was unpredictably betrayed by what they supported.`,
				Enum: []string{
					// Removal reasons
					string(ACTUAL_ANIMAL_ATTACK),
					string(BAD_EXPLANATORY_COMMENT),
					string(DIRECT_LINK_TO_OTHER_SUBREDDIT),
					string(DISTINCT_ENABLER_AND_VICTIM),
					string(DOES_NOT_FIT_THE_SUBREDDIT),
					string(FUTURE_CONSEQUENCES),
					string(LEOPARD_IN_TITLE_OR_EXPLANATORY_COMMENT),
					string(NO_CONSEQUENCES),
					string(NO_EXPLANATORY_COMMENT),

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
		Required:             []string{"reason"},
		AdditionalProperties: false,
	},
}

var (
	modFunctions = []openai.FunctionDefinition{remove, approve}
	modTools     = []openai.Tool{
		{Type: openai.ToolTypeFunction, Function: &remove},
		{Type: openai.ToolTypeFunction, Function: &approve},
	}
)

const SystemMessage = "You are a very strict moderator of r/LeopardsAteMyFace. " +
	"Your task is to approve or remove posts depending on whether they fit the subreddit or not. " +
	"Do not criticize the user's actions; only approve or remove posts. " +
	"Only communicate in English. " +
	"If the user's response is empty, consider that there are no explanatory comment."
