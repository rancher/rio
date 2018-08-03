package client

const (
	MatchType         = "match"
	MatchFieldCookies = "cookies"
	MatchFieldFrom    = "from"
	MatchFieldHeaders = "headers"
	MatchFieldMethod  = "method"
	MatchFieldPath    = "path"
	MatchFieldPort    = "port"
	MatchFieldScheme  = "scheme"
)

type Match struct {
	Cookies map[string]StringMatch `json:"cookies,omitempty" yaml:"cookies,omitempty"`
	From    *ServiceSource         `json:"from,omitempty" yaml:"from,omitempty"`
	Headers map[string]StringMatch `json:"headers,omitempty" yaml:"headers,omitempty"`
	Method  *StringMatch           `json:"method,omitempty" yaml:"method,omitempty"`
	Path    *StringMatch           `json:"path,omitempty" yaml:"path,omitempty"`
	Port    int64                  `json:"port,omitempty" yaml:"port,omitempty"`
	Scheme  *StringMatch           `json:"scheme,omitempty" yaml:"scheme,omitempty"`
}
