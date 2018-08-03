package client

const (
	WeightedDestinationType          = "weightedDestination"
	WeightedDestinationFieldPort     = "port"
	WeightedDestinationFieldRevision = "revision"
	WeightedDestinationFieldService  = "service"
	WeightedDestinationFieldStack    = "stack"
	WeightedDestinationFieldWeight   = "weight"
)

type WeightedDestination struct {
	Port     int64  `json:"port,omitempty" yaml:"port,omitempty"`
	Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`
	Service  string `json:"service,omitempty" yaml:"service,omitempty"`
	Stack    string `json:"stack,omitempty" yaml:"stack,omitempty"`
	Weight   int64  `json:"weight,omitempty" yaml:"weight,omitempty"`
}
