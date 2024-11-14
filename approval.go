package main

import "encoding/json"

func UnmarshalApproval(data []byte) (Approval, error) {
	var r Approval
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Approval) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type Approval struct {
	Someone      string `json:"someone"`
	Something    string `json:"something"`
	Consequences string `json:"consequences"`
}

type ApprovalReason string

// Event flairs were used to filter out posts that were too frequent on the
// page.
const (
	BREXXIT    ApprovalReason = "brexxit"
	COVID_19   ApprovalReason = "covid_19"
	HEALTHCARE ApprovalReason = "healthcare"
	TRUMP      ApprovalReason = "trump"
)

const (
	PREDICTABLE_BETRAYAL ApprovalReason = "predictable_betrayal"
	RISKY_BEHAVIOUR      ApprovalReason = "risky_behaviour"
)
