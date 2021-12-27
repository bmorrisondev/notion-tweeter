package main

type QueryFilter struct {
	Filter *Filter `json:"filter,omitempty"`
	Sorts  *[]Sort `json:"sorts,omitempty"`
}

type Filter struct {
	Property *string         `json:"property,omitempty"`
	Date     *DateFilter     `json:"date,omitempty"`
	Checkbox *CheckboxFilter `json:"checkbox,omitempty"`
	And      *[]Filter       `json:"and,omitempty"`
}

type DateFilter struct {
	Before *string `json:"before,omitempty"`
	After  *string `json:"after,omitempty"`
}

type CheckboxFilter struct {
	Equals    *bool `json:"equals,omitempty"`
	NotEquals *bool `json:"not_equals,omitempty"`
}

type Sort struct {
	Property  *string `json:"property,omitempty"`
	Direction *string `json:"direction,omitempty"`
}
