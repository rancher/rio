package client

const (
	AbortType             = "abort"
	AbortFieldGRPCStatus  = "grpcStatus"
	AbortFieldHTTP2Status = "http2Status"
	AbortFieldHTTPStatus  = "httpStatus"
)

type Abort struct {
	GRPCStatus  string `json:"grpcStatus,omitempty" yaml:"grpcStatus,omitempty"`
	HTTP2Status string `json:"http2Status,omitempty" yaml:"http2Status,omitempty"`
	HTTPStatus  int64  `json:"httpStatus,omitempty" yaml:"httpStatus,omitempty"`
}
