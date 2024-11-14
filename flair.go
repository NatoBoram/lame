package main

import (
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var flair = openai.FunctionDefinition{
	Name:        "flair",
	Description: "Give a flair to a post and it will be approved or removed depending on if it fits the theme of the subreddit. The flair is required.",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"flair": {
				Type: jsonschema.String,
				Description: `actual_animal_attack: The post is about an animal attacking a human
bad_explanatory_comment: It is impossible to identify who supported something or what they supported or what are the consequences from the explanatory comment
direct_link_to_other_subreddit: Contains a reference to another subreddit
distinct_enabler_and_victim: The person who supported something is not the same person as the one who receives the consequences
does_not_fit_the_subreddit: The post is not about someone who's suffering consequences from something they voted for or supported or wanted to impose on other people.
future_consequences: The consequences have not happened yet or are likely to happen
leopard_in_title_or_explanatory_comment: The words "leopards", "ate" and "face" are forbidden in the title, body and explanatory comment
no_explanatory_comment: The explanatory comment is empty
uncivil_behaviour: The user is uncivil

bye_bye_job: Someone did something and lost their job as a consequence, but losing their job isn't necessarily a consequence of what they did
hypocrisy: Someone is being a hypocrite but they're not feeling any consequences of what they supported
lesser_of_two_evils: Someone voted for something terrible, but that's only because the other choice was something even worse
self_aware_wolf: Someone accidentally describes themselves but they're not self-aware enough to realize it
stupidity: Someone is being stupid, but there's no schadenfreude to be had
sudden_betrayal: Someone was unpredictably betrayed by that they supported

brexxit: Someone voted for Brexxit and was directly impacted by the result
covid_19: Someone downplayed COVID-19 or the vaccine and got sick
healthcare: Someone voted against universal public healthcare and are now in a situation where they desperately need it
trump: Someone voted for Trump and he directly did something against that someone

predictable_betrayal: The person who was supported by someone is a known betrayer
risky_behaviour: Someone did something risky and suffered the consequences of that risk
`,
				Enum: []string{
					// Removal reasons
					string(ACTUAL_ANIMAL_ATTACK),
					string(BAD_EXPLANATORY_COMMENT),
					string(DIRECT_LINK_TO_OTHER_SUBREDDIT),
					string(DISTINCT_ENABLER_AND_VICTIM),
					string(DOES_NOT_FIT_THE_SUBREDDIT),
					string(FUTURE_CONSEQUENCES),
					string(LEOPARD_IN_TITLE_OR_EXPLANATORY_COMMENT),
					string(NO_EXPLANATORY_COMMENT),
					string(UNCIVIL_BEHAVIOUR),

					// Trapped flairs
					string(BYE_BYE_JOB),
					string(HYPOCRISY),
					string(LESSER_OF_TWO_EVILS),
					string(SELF_AWARE_WOLF),
					string(STUPIDITY),
					string(SUDDEN_BETRAYAL),

					// Event flairs
					string(BREXXIT),
					string(COVID_19),
					string(HEALTHCARE),
					string(TRUMP),

					// Approval flairs
					string(PREDICTABLE_BETRAYAL),
					string(RISKY_BEHAVIOUR),
				},
			},
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
		Required: []string{"flair", "someone", "something", "consequences"},
	},
}

var flairFunctions = []openai.FunctionDefinition{flair}
var flairTools = []openai.Tool{
	{Type: openai.ToolTypeFunction, Function: &flair},
}

func UnmarshalFlairCall(data []byte) (FlairCall, error) {
	var r FlairCall
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *FlairCall) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type FlairCall struct {
	Someone      string `json:"someone"`
	Something    string `json:"something"`
	Consequences string `json:"consequences"`
	Flair        string `json:"flair"`
}

func flairCall(resp openai.ChatCompletionResponse) (*FlairCall, error) {
	name := toolName(resp)
	if name == nil {
		return nil, fmt.Errorf("the flair tool was not called")
	}

	function := resp.Choices[0].Message.ToolCalls[0].Function
	calledFlair, err := UnmarshalFlairCall([]byte(function.Arguments))
	if err != nil {
		return &calledFlair, fmt.Errorf("failed to unmarshal flair call: %w", err)
	}

	return &calledFlair, nil
}

func suggestFlair(resp openai.ChatCompletionResponse) (*Approval, *Removal, error) {
	calledFlair, err := flairCall(resp)
	if err != nil {
		fmt.Printf("Failed to get flair call: %v\n", err)
		return nil, nil, err
	}
	decision, err := decideFlair(*calledFlair)
	if err != nil {
		fmt.Printf("Failed to decide on a flair: %v\n", err)
	}

	switch decision {
	case APPROVE:
		approval := Approval{
			Consequences: calledFlair.Consequences,
			Someone:      calledFlair.Someone,
			Something:    calledFlair.Something,
		}
		suggestApprove(approval)
		return &approval, nil, err

	case REMOVE:
		removal := Removal{
			Reason: RemovalReason(calledFlair.Flair),
		}
		suggestRemove(removal)
		return nil, &removal, err
	}

	return nil, nil, fmt.Errorf("invalid decision: %s", decision)
}

type FlairDecision string

const (
	APPROVE FlairDecision = "approve"
	REMOVE  FlairDecision = "remove"
)

func decideFlair(calledFlair FlairCall) (FlairDecision, error) {
	switch calledFlair.Flair {

	case string(ACTUAL_ANIMAL_ATTACK):
		return REMOVE, nil
	case string(BAD_EXPLANATORY_COMMENT):
		return REMOVE, nil
	case string(DIRECT_LINK_TO_OTHER_SUBREDDIT):
		return REMOVE, nil
	case string(DISTINCT_ENABLER_AND_VICTIM):
		return REMOVE, nil
	case string(DOES_NOT_FIT_THE_SUBREDDIT):
		return REMOVE, nil
	case string(FUTURE_CONSEQUENCES):
		return REMOVE, nil
	case string(LEOPARD_IN_TITLE_OR_EXPLANATORY_COMMENT):
		return REMOVE, nil
	case string(NO_EXPLANATORY_COMMENT):
		return REMOVE, nil
	case string(UNCIVIL_BEHAVIOUR):
		return REMOVE, nil
	case string(BYE_BYE_JOB):
		return REMOVE, nil
	case string(HYPOCRISY):
		return REMOVE, nil
	case string(LESSER_OF_TWO_EVILS):
		return REMOVE, nil
	case string(SELF_AWARE_WOLF):
		return REMOVE, nil
	case string(STUPIDITY):
		return REMOVE, nil
	case string(SUDDEN_BETRAYAL):
		return REMOVE, nil

	case string(BREXXIT):
		return APPROVE, nil
	case string(COVID_19):
		return APPROVE, nil
	case string(HEALTHCARE):
		return APPROVE, nil
	case string(TRUMP):
		return APPROVE, nil
	case string(PREDICTABLE_BETRAYAL):
		return APPROVE, nil
	case string(RISKY_BEHAVIOUR):
		return APPROVE, nil
	}

	return APPROVE, fmt.Errorf("invalid flair: %s", calledFlair.Flair)
}
