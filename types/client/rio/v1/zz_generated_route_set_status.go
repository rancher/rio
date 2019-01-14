package client

const (
	RouteSetStatusType            = "routeSetStatus"
	RouteSetStatusFieldConditions = "conditions"
)

type RouteSetStatus struct {
	Conditions []Condition `json:"conditions,omitempty" yaml:"conditions,omitempty"`
}
