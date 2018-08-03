package v1beta1

import (
	"strconv"
	"strings"

	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RouteSet struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RouteSetSpec `json:"spec,omitempty"`
}

type RouteSetSpec struct {
	Routes []RouteSpec `json:"routes,omitempty"`
	StackScoped
}

type RouteSpec struct {
	Matches    []Match               `json:"matches,omitempty"`
	To         []WeightedDestination `json:"to,omitempty"`
	Redirect   Redirect              `json:"redirect,omitempty"`
	Rewrite    Rewrite               `json:"rewrite,omitempty"`
	Websocket  bool                  `json:"websocket,omitempty"`
	AddHeaders []string              `json:"addHeaders,omitempty"`

	RouteTraffic
}

type RouteTraffic struct {
	Fault         Fault       `json:"fault,omitempty"`
	Mirror        Destination `json:"mirror,omitempty"`
	TimeoutMillis int         `json:"timeoutMillis,omitempty"`
	Retry         Retry       `json:"retry,omitempty"`
}

type Retry struct {
	Attempts      int `json:"attempts,omitempty"`
	TimeoutMillis int `json:"timeoutMillis,omitempty"`
}

type WeightedDestination struct {
	Destination
	Weight int64 `json:"weight,omitempty"`
}

type Destination struct {
	Service  string `json:"service,omitempty"`
	Stack    string `json:"stack,omitempty"`
	Revision string `json:"revision,omitempty"`
	Port     int64  `json:"port,omitempty"`
}

type ServiceSource struct {
	Service  string `json:"service,omitempty"`
	Stack    string `json:"stack,omitempty"`
	Revision string `json:"revision,omitempty"`
}

func (s ServiceSource) String() string {
	return Destination{
		Stack:    s.Stack,
		Service:  s.Service,
		Revision: s.Revision,
	}.String()
}

func (d Destination) String() string {
	result := strings.Builder{}
	if d.Stack != "" {
		result.WriteString(d.Stack)
		result.WriteString("/")
	}
	result.WriteString(d.Service)
	if d.Revision != "" && d.Revision != "latest" {
		result.WriteString(":")
		result.WriteString(d.Revision)
	}

	if d.Port > 0 {
		result.WriteString(",port=")
		result.WriteString(strconv.FormatInt(d.Port, 10))
	}

	return result.String()
}

func (w WeightedDestination) String() string {
	str := w.Destination.String()

	if w.Weight <= 0 {
		return str
	}

	return str + ",weight=" + strconv.FormatInt(w.Weight, 10)
}

type Fault struct {
	Percentage  int   `json:"percentage,omitempty" norman:"min=0,max=100"`
	DelayMillis int   `json:"delayMillis,omitempty"`
	Abort       Abort `json:"abort,omitempty"`
}

type Abort struct {
	HTTPStatus  int    `json:"httpStatus,omitempty"`
	HTTP2Status string `json:"http2Status,omitempty"`
	GRPCStatus  string `json:"grpcStatus,omitempty"`
}

type Match struct {
	Path    StringMatch            `json:"path,omitempty"`
	Scheme  StringMatch            `json:"scheme,omitempty"`
	Method  StringMatch            `json:"method,omitempty"`
	Headers map[string]StringMatch `json:"headers,omitempty"`
	Cookies map[string]StringMatch `json:"cookies,omitempty"`
	Port    int                    `json:"port,omitempty"`
	From    ServiceSource          `json:"from,omitempty"`
}

func (m Match) MaybeString() interface{} {
	path := m.Path.String()
	scheme := m.Scheme.String()
	method := m.Scheme.String()
	authority := m.Scheme.String()
	from := m.From.String()

	if containsComma(authority, from, method, path, scheme) ||
		containsCommaInMaps(m.Cookies, m.Headers) {
		v, _ := convert.EncodeToMap(m)
		return v
	}

	prefixData := strings.Builder{}

	addPrefixedMap(&prefixData, "cookie", m.Cookies)
	addPrefixedMap(&prefixData, "header", m.Headers)

	matchData := strings.Builder{}

	if method != "" {
		matchData.WriteString(method)
		matchData.WriteString(" ")
	}

	if scheme != "" {
		matchData.WriteString(scheme)
		matchData.WriteString("://")
	}

	matchData.WriteString(authority)

	if m.Port != 0 {
		matchData.WriteString(":")
		matchData.WriteString(strconv.Itoa(m.Port))
	}

	if len(path) > 0 && path[0] != '/' {
		matchData.WriteString("/")
	}
	matchData.WriteString(path)

	if matchData.Len() == 0 {
		return prefixData.String()
	} else if prefixData.Len() == 0 {
		return matchData.String()
	}

	prefixData.WriteString(",")
	prefixData.Write([]byte(matchData.String()))

	return prefixData.String()
}

type Redirect struct {
	Host string `json:"host,omitempty"`
	Path string `json:"path,omitempty"`
}

type Rewrite struct {
	Host string `json:"host,omitempty"`
	Path string `json:"path,omitempty"`
}

type StringMatch struct {
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Regexp string `json:"regexp,omitempty"`
}

func (s StringMatch) String() string {
	var result string
	if s.Exact != "" {
		result = s.Exact
	} else if s.Prefix != "" {
		result = s.Prefix + "*"
	} else if s.Regexp != "" {
		result = "regex(" + s.Regexp + ")"
	}

	return result
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
