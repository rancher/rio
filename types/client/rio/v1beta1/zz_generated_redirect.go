package client

const (
	RedirectType      = "redirect"
	RedirectFieldHost = "host"
	RedirectFieldPath = "path"
)

type Redirect struct {
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}
