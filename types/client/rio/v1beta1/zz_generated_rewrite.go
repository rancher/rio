package client

const (
	RewriteType      = "rewrite"
	RewriteFieldHost = "host"
	RewriteFieldPath = "path"
)

type Rewrite struct {
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}
