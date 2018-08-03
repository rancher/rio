package client

const (
	RouteSpecType               = "routeSpec"
	RouteSpecFieldAddHeaders    = "addHeaders"
	RouteSpecFieldFault         = "fault"
	RouteSpecFieldMatches       = "matches"
	RouteSpecFieldMirror        = "mirror"
	RouteSpecFieldRedirect      = "redirect"
	RouteSpecFieldRetry         = "retry"
	RouteSpecFieldRewrite       = "rewrite"
	RouteSpecFieldTimeoutMillis = "timeoutMillis"
	RouteSpecFieldTo            = "to"
	RouteSpecFieldWebsocket     = "websocket"
)

type RouteSpec struct {
	AddHeaders    []string              `json:"addHeaders,omitempty" yaml:"addHeaders,omitempty"`
	Fault         *Fault                `json:"fault,omitempty" yaml:"fault,omitempty"`
	Matches       []Match               `json:"matches,omitempty" yaml:"matches,omitempty"`
	Mirror        *Destination          `json:"mirror,omitempty" yaml:"mirror,omitempty"`
	Redirect      *Redirect             `json:"redirect,omitempty" yaml:"redirect,omitempty"`
	Retry         *Retry                `json:"retry,omitempty" yaml:"retry,omitempty"`
	Rewrite       *Rewrite              `json:"rewrite,omitempty" yaml:"rewrite,omitempty"`
	TimeoutMillis int64                 `json:"timeoutMillis,omitempty" yaml:"timeoutMillis,omitempty"`
	To            []WeightedDestination `json:"to,omitempty" yaml:"to,omitempty"`
	Websocket     bool                  `json:"websocket,omitempty" yaml:"websocket,omitempty"`
}
