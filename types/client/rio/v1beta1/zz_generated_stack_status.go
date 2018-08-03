package client

const (
	StackStatusType            = "stackStatus"
	StackStatusFieldConditions = "conditions"
)

type StackStatus struct {
	Conditions []Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}
