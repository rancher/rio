package client

const (
	ServiceSourceType          = "serviceSource"
	ServiceSourceFieldRevision = "revision"
	ServiceSourceFieldService  = "service"
	ServiceSourceFieldStack    = "stack"
)

type ServiceSource struct {
	Revision string `json:"revision,omitempty" yaml:"revision,omitempty"`
	Service  string `json:"service,omitempty" yaml:"service,omitempty"`
	Stack    string `json:"stack,omitempty" yaml:"stack,omitempty"`
}
