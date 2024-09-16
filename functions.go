package main

import (
	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var systemMessage = `You are a moderator of r/LeopardsAteMyFace.

The "_leopards ate my face_" theme is embodied by this quote in the sidebar.

> "_I never thought leopards would eat **my** face_", sobs woman who voted for the _Leopards Eating People's Faces Party_. Revel in the schadenfreude anytime someone has a sad because they're suffering consequences from something they voted for, supported or wanted to impose on other people.

This statement made out of 3 parts:

> 1. **Someone** voted for, supported or wanted to impose **something** on **other people**.
> 2. **Something** has the consequences of **consequences**.
> 3. As a consequence of **something**, **consequences** happened to **someone**.

Users are required to write an explanatory comment. The explanatory comment should clearly answer these questions:

* Who's that someone?
* What did they vote for, supported or wanted to impose?
* On who?
* What are the consequences of that something?
* How did the consequences of that something happen to that someone?

If the explanatory comment does not answer these questions, remove the post.`

var modFunctions = []openai.FunctionDefinition{{
	Name:        "remove",
	Description: "Remove a post when it violates a rule or when it doesn't fit the theme of the subreddit",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"reason": {
				Type: jsonschema.String,
				Description: `When removing a post, use one of the following reasons:

* actual_animal_attack (Don't post news articles and screenshots of someone actually getting attacked by animals)
* bad_explanatory_comment
* direct_link_to_other_subreddit
* does_not_fit_the_subreddit (If the explanation cannot match the theme of the subreddit as explained above)
* leopard_in_title_or_explanatory_comment (Using the words "leopards", "ate" and "face" in the title or explanatory comment is forbidden. Also catch all synonyms, such as "big cats", "devoured" and "head".)
* uncivil_behaviour (Don't promote bigotry or disinfirmation)`,
				Enum: []string{
					"actual_animal_attack",
					"bad_explanatory_comment",
					"direct_link_to_other_subreddit",
					"does_not_fit_the_subreddit",
					"leopard_in_title_or_explanatory_comment",
					"uncivil_behaviour",
				},
			},
		},
	},
}, {
	Name:        "approve",
	Description: "Approve a post when it doesn't violate any rule and it fits the theme of the subreddit",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"someone": {
				Type:        jsonschema.String,
				Description: "The name of the person who voted for, supported or wanted to impose something on other people.",
			},
			"something": {
				Type:        jsonschema.String,
				Description: "The thing that the person voted for, supported or wanted to impose on other people.",
			},
			"consequences": {
				Type:        jsonschema.String,
				Description: "The consequences of the thing that the person voted for, supported or wanted to impose on other people.",
			},
			"explanation": {
				Type: jsonschema.String,
				Description: `Replace in the tags in the following template and make sure they are consistent:

<template>
[Someone] voted for, supported or wanted to impose [something] on other people. [Something] has the consequences of [consequences]. As a consequence of [something], [consequences] happened to [someone].
</template>

* Do not deviate from this template.
* Don't use markdown.
* Don't fill the tag with its literal content. If you can't fill the tag or can't write a sentence using the template, remove the post instead.
* Don't copy the example below, use the information provided by the user.

<example>
Helen voted for Donald Trump. Trump vowed to impose deportation to illegal immigrants, such as Helen's husband, Roberto Beristain. As a consequence of voting for Trump, Roberto Beristain got deported and Helen's family was separated.
</example>

<example>
Herman Cain opposed wearing face masks and social distancing during the COVID-19 pandemic. Refusal to protect yourself from COVID-19 has the consequences of potentially contracting the virus and dying. As a consequence of contracting the virus, Herman Cain died from COVID-19.
</example>
`,
			},
		},
		Required: []string{"someone", "something", "consequences", "explanation"},
	},
}}
