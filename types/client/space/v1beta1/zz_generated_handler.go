package client

const (
	HandlerType           = "handler"
	HandlerFieldExec      = "exec"
	HandlerFieldHTTPGet   = "httpGet"
	HandlerFieldTCPSocket = "tcpSocket"
)

type Handler struct {
	Exec      *ExecAction      `json:"exec,omitempty" yaml:"exec,omitempty"`
	HTTPGet   *HTTPGetAction   `json:"httpGet,omitempty" yaml:"httpGet,omitempty"`
	TCPSocket *TCPSocketAction `json:"tcpSocket,omitempty" yaml:"tcpSocket,omitempty"`
}
