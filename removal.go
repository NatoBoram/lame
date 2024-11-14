package main

import "encoding/json"

func UnmarshalRemoval(data []byte) (Removal, error) {
	var r Removal
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Removal) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Removal struct {
	Reason RemovalReason `json:"reason"`
}

type RemovalReason string

const (
	ACTUAL_ANIMAL_ATTACK                    RemovalReason = "actual_animal_attack"
	BAD_EXPLANATORY_COMMENT                 RemovalReason = "bad_explanatory_comment"
	DIRECT_LINK_TO_OTHER_SUBREDDIT          RemovalReason = "direct_link_to_other_subreddit"
	DOES_NOT_FIT_THE_SUBREDDIT              RemovalReason = "does_not_fit_the_subreddit"
	LEOPARD_IN_TITLE_OR_EXPLANATORY_COMMENT RemovalReason = "leopard_in_title_or_explanatory_comment"
	NO_CONSEQUENCES_YET                     RemovalReason = "no_consequences_yet"
	NO_EXPLANATORY_COMMENT                  RemovalReason = "no_explanatory_comment"
	UNCIVIL_BEHAVIOUR                       RemovalReason = "uncivil_behaviour"
)

// Trapped flairs were used to easily identify kinds of posts that were not
// allowed.
const (
	BYE_BYE_JOB         RemovalReason = "bye_bye_job"
	HYPOCRISY           RemovalReason = "hypocrisy"
	LESSER_OF_TWO_EVILS RemovalReason = "lesser_of_two_evils"
	SELF_AWARE_WOLF     RemovalReason = "self_aware_wolf"
	STUPIDITY           RemovalReason = "stupidity"
	SUDDEN_BETRAYAL     RemovalReason = "sudden_betrayal"
)

var trappedFlairs = []RemovalReason{
	BYE_BYE_JOB,
	HYPOCRISY,
	LESSER_OF_TWO_EVILS,
	SELF_AWARE_WOLF,
	STUPIDITY,
	SUDDEN_BETRAYAL,
}
