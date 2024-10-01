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
	UNCIVIL_BEHAVIOUR                       RemovalReason = "uncivil_behaviour"
)
