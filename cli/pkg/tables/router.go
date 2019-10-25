package tables

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/rancher/rio/cli/pkg/table"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type RouteSpecData struct {
	Name      string
	Obj       *v1.Router
	RouteSpec v1.RouteSpec
	Match     *v1.Match
	Namespace string
	Domain    string
}

func (d *RouteSpecData) path() string {
	if d.Match == nil {
		return ""
	}
	if d.Match.Path == nil {
		return ""
	}
	str := ""
	if d.Match.Path != nil {
		str = d.Match.Path.String()
	}
	if str != "" && str[0] != '/' {
		return "/" + str
	}
	return str
}

func stringMatchToString(m *v1.StringMatch) string {
	if m == nil {
		return ""
	}
	return m.String()
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

	for j := range routeSet.Spec.Routes {
		r.Writer.Write(&RouteSpecData{
			Obj:       routeSet,
			RouteSpec: routeSet.Spec.Routes[j],
			Match:     &routeSet.Spec.Routes[j].Match,
			Domain:    r.domain,
			Namespace: routeSet.Namespace,
		})
	}
}

func NewRouter(cfg Config) TableWriter {
	writer := table.NewWriter([][]string{
		{"NAME", "{{id .Obj}}"},
		{"URL", "{{ . | formatURL }}"},
		{"OPTS", "{{ . | formatOpts }}"},
		{"ACTION", "{{ . | formatAction }}"},
		{"TARGET", "{{ . | formatTarget }}"},
	}, cfg)

	writer.AddFormatFunc("formatOpts", FormatOpts)
	writer.AddFormatFunc("formatURL", FormatURL())
	writer.AddFormatFunc("formatAction", FormatAction)
	writer.AddFormatFunc("formatTarget", FormatRouteTarget)

	domain, _ := cfg.Domain()

	return &tableWriter{
		writer: &routerWriter{
			Writer: writer,
			domain: domain.Name,
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

	if data.RouteSpec.Mirror != nil && data.RouteSpec.Mirror.App != "" {
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
		writeStringMatchMap(buf, "header=", data.Match.Headers)
		writeStringMatchForList(buf, "method=", data.Match.Methods)
	}

	return buf.String(), nil
}

func FormatRouteTarget(obj interface{}) (string, error) {
	buf := &strings.Builder{}
	data, ok := obj.(*RouteSpecData)
	if !ok {
		return "", fmt.Errorf("invalid data")
	}

	switch target := targetType(data); {
	case target == "to":
		for _, to := range data.RouteSpec.To {
			if to.Port == 0 {
				to.Port = uint32(80)
			}
			writeDest(buf, data.Obj.Namespace, to.App, to.Version, to.Port, to.Weight)
		}
	case target == "redirect" && data.RouteSpec.Redirect != nil:
		writeHostPath(buf, data.RouteSpec.Redirect.Host, data.RouteSpec.Redirect.Path)
	case target == "mirror" && data.RouteSpec.Mirror != nil:
		if data.RouteSpec.Mirror.Port == 0 {
			data.RouteSpec.Mirror.Port = 80
		}
		writeDest(buf, data.Obj.Namespace, data.RouteSpec.Mirror.App,
			data.RouteSpec.Mirror.Version,
			data.RouteSpec.Mirror.Port, 0)
	case target == "rewrite" && data.RouteSpec.Rewrite != nil:
		writeHostPath(buf, data.RouteSpec.Rewrite.Host, data.RouteSpec.Rewrite.Path)
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

func writeDest(buf *strings.Builder, namespace, service, revision string, port uint32, weight int) {
	if buf.Len() > 0 {
		buf.WriteString(",")
	}

	buf.WriteString(namespace)
	buf.WriteString("/")

	buf.WriteString(service)
	if revision != "" && revision != "latest" {
		buf.WriteString(":")
		buf.WriteString(revision)
	}

	if port > 0 {
		buf.WriteString(",port=")
		buf.WriteString(strconv.Itoa(int(port)))
	}

	if weight > 0 {
		buf.WriteString(" ")
		buf.WriteString(strconv.Itoa(weight))
		buf.WriteString("%")
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

func writeStringMatchForList(buf *strings.Builder, prefix string, values []string) {
	if buf.Len() > 0 {
		buf.WriteString(",")
	}
	for _, v := range values {
		buf.WriteString(prefix)
		buf.WriteString("=")
		buf.WriteString(v)
	}

}

func writeStringMatchMap(buf *strings.Builder, prefix string, matches []v1.HeaderMatch) {
	for _, m := range matches {
		writeStringMatch(buf, prefix+m.Name, m.Value)
	}
}

func FormatURL() func(obj interface{}) (string, error) {
	return func(obj interface{}) (string, error) {
		data, ok := obj.(*RouteSpecData)
		if !ok {
			return "", fmt.Errorf("invalid data")
		}
		if len(data.Obj.Status.Endpoints) == 0 {
			return "", nil
		}
		var endpoints []string
		hostNameSeen := map[string]string{}
		for _, e := range data.Obj.Status.Endpoints {
			u, _ := url.Parse(e)
			if u.Scheme == "https" {
				hostNameSeen[u.Hostname()] = e + data.path()
			} else {
				if _, ok := hostNameSeen[u.Hostname()]; !ok {
					hostNameSeen[u.Hostname()] = e + data.path()
				}
			}
		}

		for _, endpoint := range hostNameSeen {
			endpoints = append(endpoints, endpoint)
		}
		return strings.Join(endpoints, ","), nil
	}
}
