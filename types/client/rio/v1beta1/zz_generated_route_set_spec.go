package client

const (
	RouteSetSpecType         = "routeSetSpec"
	RouteSetSpecFieldRoutes  = "routes"
	RouteSetSpecFieldSpaceID = "spaceId"
	RouteSetSpecFieldStackID = "stackId"
)

type RouteSetSpec struct {
	Routes  []RouteSpec `json:"routes,omitempty" yaml:"routes,omitempty"`
	SpaceID string      `json:"spaceId,omitempty" yaml:"spaceId,omitempty"`
	StackID string      `json:"stackId,omitempty" yaml:"stackId,omitempty"`
}
