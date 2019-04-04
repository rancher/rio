package tables

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/rio/cli/pkg/table"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/name"
)

type RouteSpecData struct {
	ID        string
	RouteSet  v1.Router
	RouteSpec v1.RouteSpec
	Match     *v1.Match
	Domain    string
}

func (d *RouteSpecData) port() int {
	if d.Match == nil || d.Match.Port == nil {
		return 0
	}
	return *d.Match.Port
}

func (d *RouteSpecData) path() string {
	if d.Match == nil {
		return ""
	}
	if d.Match.Path == nil {
		return ""
	}
	str := stringMatchToString(d.Match.Path)
	if str != "" && str[0] != '/' {
		return "/" + str
	}
	return str
}

func stringMatchToString(m *v1.StringMatch) string {
	if m == nil {
		return ""
	}
	return v1.StringMatch{
		Exact:  m.Exact,
		Regexp: m.Regexp,
		Prefix: m.Prefix,
	}.String()
}

type routerWriter struct {
	table.Writer
	domain string
}

func (r *routerWriter) Write(obj interface{}) {
	td, ok := obj.(*data)
	if !ok {
		return
	}

	routeSet, ok := td.Obj.(*v1.Router)
	if !ok {
		return
	}

	for j, routeSpec := range routeSet.Spec.Routes {
		if len(routeSpec.Matches) == 0 {
			r.Writer.Write(&RouteSpecData{
				ID:        routeSet.Name,
				RouteSet:  *routeSet,
				RouteSpec: routeSet.Spec.Routes[j],
				Domain:    r.domain,
			})
			continue
		}

		for k := range routeSpec.Matches {
			r.Writer.Write(&RouteSpecData{
				RouteSet:  *routeSet,
				RouteSpec: routeSet.Spec.Routes[j],
				Match:     &routeSet.Spec.Routes[j].Matches[k],
				Domain:    r.domain,
			})
		}
	}
}

func NewRouter(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Obj.Namespace .Obj.Name}}"},
		{"URL", "{{ . | formatURL }}"},
		{"OPTS", "{{ . | formatOpts }}"},
		{"ACTION", "{{ . | formatAction }}"},
		{"TARGET", "{{ . | formatTarget }}"},
	}, cfg)

	writer.AddFormatFunc("formatOpts", FormatOpts)
	writer.AddFormatFunc("formatURL", FormatURL())
	writer.AddFormatFunc("formatAction", FormatAction)
	writer.AddFormatFunc("formatTarget", FormatRouteTarget)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cfg.GetDefaultStackName()))

	domain, _ := cfg.Domain()

	return &tableWriter{
		writer: &routerWriter{
			Writer: writer,
			domain: domain,
		},
	}
}

func FormatAction(obj interface{}) (string, error) {
	data, ok := obj.(*RouteSpecData)
	if !ok {
		return "", fmt.Errorf("invalid data")
	}

	return targetType(data), nil
}

func targetType(data *RouteSpecData) string {
	if data.RouteSpec.Rewrite != nil &&
		(len(data.RouteSpec.Rewrite.Host) > 0 || len(data.RouteSpec.Rewrite.Path) > 0) {
		return "rewrite"
	}

	if data.RouteSpec.Redirect != nil &&
		(len(data.RouteSpec.Redirect.Host) > 0 || len(data.RouteSpec.Redirect.Path) > 0) {
		return "redirect"
	}

	if data.RouteSpec.Mirror != nil && len(data.RouteSpec.Mirror.Service) > 0 {
		return "mirror"
	}

	return "to"
}

func FormatOpts(obj interface{}) (string, error) {
	buf := &strings.Builder{}
	data, ok := obj.(*RouteSpecData)
	if !ok {
		return "", fmt.Errorf("invalid data")
	}

	if data.Match != nil {
		writeStringMatchMap(buf, "cookie=", data.Match.Cookies)
		writeStringMatchMap(buf, "header=", data.Match.Headers)
		addFrom(buf, data.Match.From)
		writeStringMatch(buf, "method=", data.Match.Method)
		if data.Match.Scheme != nil {
			buf.WriteString(stringMatchToString(data.Match.Scheme))
		}
	}

	if data.RouteSpec.TimeoutMillis != nil {
		if buf.Len() > 0 {
			buf.WriteString(",")
		}
		buf.WriteString("timeout=")
		d := time.Duration(*data.RouteSpec.TimeoutMillis) * time.Millisecond
		buf.WriteString(d.String())
	}

	return buf.String(), nil
}

func FormatRouteTarget(obj interface{}) (string, error) {
	buf := &strings.Builder{}
	data, ok := obj.(*RouteSpecData)
	if !ok {
		return "", fmt.Errorf("invalid data")
	}

	target := targetType(data)

	if target == "to" {
		for _, to := range data.RouteSpec.To {
			writeDest(buf, data.RouteSet.Namespace, to.Stack, to.Service, to.Revision, int(*to.Port), to.Weight)
		}
	} else if target == "redirect" && data.RouteSpec.Redirect != nil {
		writeHostPath(buf, data.RouteSpec.Redirect.Host, data.RouteSpec.Redirect.Path)
	} else if target == "mirror" && data.RouteSpec.Mirror != nil && data.RouteSpec.Mirror.Port != nil {
		writeDest(buf, data.RouteSet.Namespace, data.RouteSpec.Mirror.Stack, data.RouteSpec.Mirror.Service,
			data.RouteSpec.Mirror.Revision,
			int(*data.RouteSpec.Mirror.Port), 0)
	} else if target == "rewrite" {
		writeHostPath(buf, data.RouteSpec.Redirect.Host, data.RouteSpec.Redirect.Path)
	}

	return buf.String(), nil
}

func writeHostPath(buf *strings.Builder, host, path string) {
	if host != "" {
		buf.WriteString(host)
	}
	if path != "" {
		if path[0] != '/' {
			buf.WriteString("/")
		}
		buf.WriteString(path)
	}
}

func writeDest(buf *strings.Builder, sourceStackName, stack, service, revision string, port int, weight int) {
	if buf.Len() > 0 {
		buf.WriteString(",")
	}

	if stack != "" && stack != sourceStackName {
		buf.WriteString(stack)
		buf.WriteString("/")
	}
	buf.WriteString(service)
	if revision != "" && revision != "latest" {
		buf.WriteString(":")
		buf.WriteString(revision)
	}

	if port > 0 {
		buf.WriteString(",port=")
		buf.WriteString(strconv.Itoa(port))
	}

	if weight > 0 {
		buf.WriteString(" ")
		buf.WriteString(strconv.Itoa(weight))
		buf.WriteString("%")
	}
}

func addFrom(buf *strings.Builder, from *v1.ServiceSource) {
	if from == nil {
		return
	}

	if from.Service == "" {
		return
	}

	if buf.Len() > 0 {
		buf.WriteString(",")
	}
	buf.WriteString("from=")
	if from.Stack != "" {
		buf.WriteString(from.Stack)
		buf.WriteString("/")
	}
	buf.WriteString(from.Service)
	if from.Revision != "" {
		buf.WriteString(":")
		buf.WriteString(from.Revision)
	}
}

func writeStringMatch(buf *strings.Builder, prefix string, sm *v1.StringMatch) {
	if sm == nil {
		return
	}

	if buf.Len() > 0 {
		buf.WriteString(",")
	}
	buf.WriteString(prefix)
	str := stringMatchToString(sm)
	if str != "" {
		buf.WriteString("=")
		buf.WriteString(str)
	}
}

func writeStringMatchMap(buf *strings.Builder, prefix string, matches map[string]v1.StringMatch) {
	var keys []string
	for k := range matches {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		m := matches[k]
		writeStringMatch(buf, prefix+k, &m)
	}
}

func FormatURL() func(obj interface{}) (string, error) {
	return func(obj interface{}) (string, error) {
		data, ok := obj.(*RouteSpecData)
		if !ok {
			return "", fmt.Errorf("invalid data")
		}
		hostBuf := strings.Builder{}
		hostBuf.WriteString("https://")
		name := name.SafeConcatName(data.RouteSet.Name, data.RouteSet.Name, data.RouteSet.Namespace)
		hostBuf.WriteString(name)
		hostBuf.WriteString(".")
		hostBuf.WriteString(data.Domain)
		if data.port() > 0 {
			hostBuf.WriteString(":")
			hostBuf.WriteString(strconv.Itoa(data.port()))
		}
		hostBuf.WriteString(data.path())
		return hostBuf.String(), nil
	}
}
