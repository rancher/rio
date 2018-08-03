package client

const (
	StringMatchType        = "stringMatch"
	StringMatchFieldExact  = "exact"
	StringMatchFieldPrefix = "prefix"
	StringMatchFieldRegexp = "regexp"
)

type StringMatch struct {
	Exact  string `json:"exact,omitempty" yaml:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Regexp string `json:"regexp,omitempty" yaml:"regexp,omitempty"`
}
