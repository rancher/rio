package v1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Router is a top level resource to create L7 routing to different services. It will create VirtualService, ServiceEntry and DestinationRules
type Router struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouterSpec   `json:"spec,omitempty"`
	Status RouterStatus `json:"status,omitempty"`
}

type RouterSpec struct {
	// An ordered list of route rules for HTTP traffic. The first rule matching an incoming request is used.
	Routes []RouteSpec `json:"routes,omitempty"`

	// By default all Routers are public and exposed outside of the cluster. Setting internal to true will
	// cause the Router to not be exposed
	Internal bool `json:"internal,omitempty"`
}

type RouterStatus struct {
	// The endpoint to access the router
	Endpoints []string `json:"endpoints,omitempty" column:"name=Endpoint,type=string,jsonpath=.status.endpoints[0]"`

	// Represents the latest available observations of a PublicDomain's current state.
	Conditions []genericcondition.GenericCondition `json:"conditions,omitempty"`
}

type RouteSpec struct {
	//Match conditions to be satisfied for the rule to be activated. All conditions inside a single match block have AND semantics, while the list of match blocks have OR semantics.
	// The rule is matched if any one of the match blocks succeed.
	Match Match `json:"match,omitempty"`

	// A http rule can either redirect or forward (default) traffic. The forwarding target can be one of several versions of a service (see glossary in beginning of document).
	// Weights associated with the service version determine the proportion of traffic it receives.
	To []WeightedDestination `json:"to,omitempty"`

	// A http rule can either redirect or forward (default) traffic. If traffic passthrough option is specified in the rule, route/redirect will be ignored.
	// The redirect primitive can be used to send a HTTP 301 redirect to a different URI or Authority.
	Redirect *Redirect `json:"redirect,omitempty"`

	// Rewrite HTTP URIs and Authority headers. Rewrite cannot be used with Redirect primitive. Rewrite will be performed before forwarding.
	Rewrite *Rewrite `json:"rewrite,omitempty"`

	// Retries specifies the retry logic for each route
	Retry *Retry `json:"retry,omitempty"`

	//Header manipulation rules
	Headers *HeaderOperations `json:"headers,omitempty"`

	// Fault injection policy to apply on HTTP traffic at the client side. Note that timeouts or retries will not be enabled when faults are enabled on the client side.
	Fault *Fault `json:"fault,omitempty"`

	// Mirror HTTP traffic to a another destination in addition to forwarding the requests to the intended destination.
	// Mirrored traffic is on a best effort basis where the sidecar/gateway will not wait for the mirrored cluster to respond before returning the response from the original destination.
	// Statistics will be generated for the mirrored destination.
	Mirror *Destination `json:"mirror,omitempty"`

	// TimeoutSeconds specifies timeout setting for each route
	TimeoutSeconds *int `json:"timeoutSeconds,omitempty"`
}

// HeaderOperations Describes the header manipulations to apply
type HeaderOperations struct {
	// Append the given values to the headers specified by keys
	// (will create a comma-separated list of values)
	Add []NameValue `json:"add,omitempty"`

	// Append the given values to the headers specified by keys
	// (will create a comma-separated list of values)
	Set []NameValue `json:"set,omitempty"`

	// Remove a the specified headers
	Remove []string `json:"remove,omitempty"`
}

type NameValue struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type WeightedDestination struct {
	Destination

	// Weight for the Destination
	Weight int `json:"weight,omitempty"`
}

type Destination struct {
	// Destination Service
	App string `json:"app,omitempty"`

	// Destination Revision
	Version string `json:"version,omitempty"`

	// Destination Port
	Port uint32 `json:"port,omitempty"`
}

func (d Destination) String() string {
	result := strings.Builder{}
	result.WriteString(d.App)
	if d.Version != "" && d.Version != "latest" {
		result.WriteString(":")
		result.WriteString(d.Version)
	}

	if d.Port > 0 {
		result.WriteString(",port=")
		result.WriteString(strconv.FormatInt(int64(d.Port), 10))
	}

	return result.String()
}

func (w WeightedDestination) String() string {
	str := w.Destination.String()

	if w.Weight <= 0 {
		return str
	}

	return str + ",weight=" + strconv.FormatInt(int64(w.Weight), 10)
}

type Fault struct {
	// Percentage of requests on which the delay will be injected.
	Percentage int `json:"percentage,omitempty" norman:"min=0,max=100"`

	// REQUIRED. Add a fixed delay before forwarding the request. Units: milliseconds
	DelayMillis int `json:"delayMillis,omitempty"`

	// Abort Http request attempts and return error codes back to downstream service, giving the impression that the upstream service is faulty.
	AbortHTTPStatus int `json:"abortHTTPStatus,omitempty"`
}

type Match struct {
	//URI to match values are case-sensitive and formatted as follows:
	//
	//    exact: "value" for exact string match
	//
	//    prefix: "value" for prefix-based match
	//
	//    regex: "value" for ECMAscript style regex-based match
	Path *StringMatch `json:"path,omitempty"`

	// HTTP Method values are case-sensitive and formatted as follows:
	//
	//    exact: "value" for exact string match
	//
	//    prefix: "value" for prefix-based match
	//
	//    regex: "value" for ECMAscript style regex-based match
	Methods []string `json:"methods,omitempty"`

	// The header keys must be lowercase and use hyphen as the separator, e.g. x-request-id.
	//
	// Header values are case-sensitive and formatted as follows:
	//
	//    exact: "value" for exact string match
	//
	//    prefix: "value" for prefix-based match
	//
	//    regex: "value" for ECMAscript style regex-based match
	//
	// Note: The keys uri, scheme, method, and authority will be ignored.
	Headers []HeaderMatch `json:"headers,omitempty"`
}

type HeaderMatch struct {
	Name  string       `json:"name,omitempty"`
	Value *StringMatch `json:"value,omitempty"`
}

func (h HeaderMatch) String() string {
	value := ""
	if h.Value != nil {
		value = h.Value.String()
	}
	return fmt.Sprintf("%s=%s", h.Name, value)
}

func (m Match) MaybeString() interface{} {
	return ""
	//path := m.Path.String()
	//method := m.Methods

	//if containsComma(authority, from, method, path, scheme) ||
	//	containsCommaInMaps(m.Cookies, m.Headers) {
	//	v, _ := convert.EncodeToMap(m)
	//	return v
	//}
	//
	//prefixData := strings.Builder{}
	//
	//addPrefixedMap(&prefixData, "cookie", m.Cookies)
	//addPrefixedMap(&prefixData, "header", m.Headers)
	//
	//matchData := strings.Builder{}
	//
	//if method != "" {
	//	matchData.WriteString(method)
	//	matchData.WriteString(" ")
	//}
	//
	//if scheme != "" {
	//	matchData.WriteString(scheme)
	//	matchData.WriteString("://")
	//}
	//
	//matchData.WriteString(authority)
	//
	//if m.Port != nil && *m.Port != 0 {
	//	matchData.WriteString(":")
	//	matchData.WriteString(strconv.Itoa(*m.Port))
	//}
	//
	//if len(path) > 0 && path[0] != '/' {
	//	matchData.WriteString("/")
	//}
	//matchData.WriteString(path)
	//
	//if matchData.Len() == 0 {
	//	return prefixData.String()
	//} else if prefixData.Len() == 0 {
	//	return matchData.String()
	//}
	//
	//prefixData.WriteString(",")
	//prefixData.Write([]byte(matchData.String()))
	//
	//return prefixData.String()
}

type Redirect struct {
	Host    string `json:"host,omitempty"`
	Path    string `json:"path,omitempty"`
	Prefix  string `json:"prefix,omitempty"`
	ToHTTPS bool   `json:"toHTTPS,omitempty"`
}

type Rewrite struct {
	Host string `json:"host,omitempty"`
	Path string `json:"path,omitempty"`
}

type Retry struct {
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`
	Attempts       int `json:"attempts,omitempty"`
}

type StringMatch struct {
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Regexp string `json:"regexp,omitempty"`
}

func (s StringMatch) String() string {
	switch {
	case s.Exact != "":
		return s.Exact
	case s.Prefix != "":
		return s.Prefix + "*"
	case s.Regexp != "":
		return "regex(" + s.Regexp + ")"
	default:
		return ""
	}
}

func containsComma(strs ...string) bool {
	for _, str := range strs {
		if strings.ContainsRune(str, ',') {
			return true
		}
	}
	return false
}

func containsCommaInMaps(maps ...map[string]StringMatch) bool {
	for _, m := range maps {
		for k, v := range m {
			if strings.ContainsRune(k, ',') {
				return true
			}
			if strings.ContainsRune(v.String(), ',') {
				return true
			}
		}
	}
	return false
}

func addPrefixedMap(buf *strings.Builder, prefix string, data map[string]StringMatch) {
	for k, v := range data {
		if buf.Len() > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(prefix)
		buf.WriteString("=")
		buf.WriteString(k)

		str := v.String()
		if str != "" {
			buf.WriteString("=")
			buf.WriteString(str)
		}
	}
}
