package client

const (
	ServerType       = "server"
	ServerFieldHosts = "hosts"
	ServerFieldPort  = "port"
)

type Server struct {
	Hosts []string `json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Port  *Port    `json:"port,omitempty" yaml:"port,omitempty"`
}
