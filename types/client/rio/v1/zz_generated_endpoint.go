package client

const (
	EndpointType     = "endpoint"
	EndpointFieldURL = "url"
)

type Endpoint struct {
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
}
