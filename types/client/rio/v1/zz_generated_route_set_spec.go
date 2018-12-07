package client

const (
	RouteSetSpecType           = "routeSetSpec"
	RouteSetSpecFieldProjectID = "projectId"
	RouteSetSpecFieldRoutes    = "routes"
	RouteSetSpecFieldStackID   = "stackId"
)

type RouteSetSpec struct {
	ProjectID string      `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	Routes    []RouteSpec `json:"routes,omitempty" yaml:"routes,omitempty"`
	StackID   string      `json:"stackId,omitempty" yaml:"stackId,omitempty"`
}
