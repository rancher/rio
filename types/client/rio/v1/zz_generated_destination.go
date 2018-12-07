package client

const (
	DestinationType          = "destination"
	DestinationFieldPort     = "port"
	DestinationFieldRevision = "revision"
	DestinationFieldService  = "service"
	DestinationFieldStack    = "stack"
)

type Destination struct {
	Port     int64  `json:"port,omitempty" yaml:"port,omitempty"`
	Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`
	Service  string `json:"service,omitempty" yaml:"service,omitempty"`
	Stack    string `json:"stack,omitempty" yaml:"stack,omitempty"`
}
