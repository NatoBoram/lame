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
	Explanation string `json:"explanation"`
}
