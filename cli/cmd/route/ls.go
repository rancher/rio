package route

import (
	"fmt"

	"strconv"
	"strings"

	"sort"

	"time"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Data struct {
	ID        string
	Stack     *client.Stack
	RouteSet  client.RouteSet
	RouteSpec client.RouteSpec
	Match     *client.Match
	Domain    string
}

func (d *Data) port() int64 {
	if d.Match == nil {
		return 0
	}
	return d.Match.Port
}

func (d *Data) path() string {
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

func stringMatchToString(m *client.StringMatch) string {
	if m == nil {
		return ""
	}
	return v1beta1.StringMatch{
		Exact:  m.Exact,
		Regexp: m.Regexp,
		Prefix: m.Prefix,
	}.String()
}

type Ls struct {
	L_Label map[string]string `desc:"Set meta data on a container"`
}

func (l *Ls) Customize(cmd *cli.Command) {
	cmd.Flags = append(cmd.Flags, table.WriterFlags()...)
}

func (l *Ls) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	routeSets, err := ctx.Client.RouteSet.List(util.DefaultListOpts())
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"URL", "{{ . | formatURL }}"},
		{"OPTS", "{{ . | formatOpts }}"},
		{"ACTION", "{{ . | formatAction }}"},
		{"TARGET", "{{ . | formatTarget }}"},
	}, app)
	defer writer.Close()

	writer.AddFormatFunc("formatURL", FormatURL)
	writer.AddFormatFunc("formatOpts", FormatOpts)
	writer.AddFormatFunc("formatAction", FormatAction)
	writer.AddFormatFunc("formatTarget", FormatTarget)

	stackByID, err := util.StacksByID(ctx)
	if err != nil {
		return err
	}

	for i, routeSet := range routeSets.Data {
		for j, routeSpec := range routeSet.Routes {
			if len(routeSpec.Matches) == 0 {
				writer.Write(&Data{
					ID:        routeSet.ID,
					RouteSet:  routeSets.Data[i],
					RouteSpec: routeSets.Data[i].Routes[j],
					Stack:     stackByID[routeSet.StackID],
					Domain:    ctx.Domain,
				})
				continue
			}

			for k := range routeSpec.Matches {
				writer.Write(&Data{
					ID:        routeSet.ID,
					RouteSet:  routeSets.Data[i],
					RouteSpec: routeSets.Data[i].Routes[j],
					Match:     &routeSets.Data[i].Routes[j].Matches[k],
					Stack:     stackByID[routeSet.StackID],
					Domain:    ctx.Domain,
				})
			}
		}
	}

	return writer.Err()
}

func FormatAction(obj interface{}) (string, error) {
	data, ok := obj.(*Data)
	if !ok {
		return "", fmt.Errorf("invalid data")
	}

	return targetType(data), nil
}

func targetType(data *Data) string {
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
	data, ok := obj.(*Data)
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

	if data.RouteSpec.TimeoutMillis > 0 {
		if buf.Len() > 0 {
			buf.WriteString(",")
		}
		buf.WriteString("timeout=")
		d := time.Duration(data.RouteSpec.TimeoutMillis) * time.Millisecond
		buf.WriteString(d.String())
	}

	return buf.String(), nil
}

func FormatTarget(obj interface{}) (string, error) {
	buf := &strings.Builder{}
	data, ok := obj.(*Data)
	if !ok {
		return "", fmt.Errorf("invalid data")
	}

	target := targetType(data)

	if target == "to" {
		for _, to := range data.RouteSpec.To {
			writeDest(buf, data.Stack.Name, to.Stack, to.Service, to.Revision, to.Port, to.Weight)
		}
	} else if target == "redirect" && data.RouteSpec.Redirect != nil {
		writeHostPath(buf, data.RouteSpec.Redirect.Host, data.RouteSpec.Redirect.Path)
	} else if target == "mirror" && data.RouteSpec.Mirror != nil {
		writeDest(buf, data.Stack.Name, data.RouteSpec.Mirror.Stack, data.RouteSpec.Mirror.Service,
			data.RouteSpec.Mirror.Revision,
			data.RouteSpec.Mirror.Port, 0)
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

func writeDest(buf *strings.Builder, sourceStackName, stack, service, revision string, port, weight int64) {
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
		buf.WriteString(strconv.FormatInt(port, 10))
	}

	if weight > 0 {
		buf.WriteString(" ")
		buf.WriteString(strconv.FormatInt(weight, 10))
		buf.WriteString("%")
	}
}

func addFrom(buf *strings.Builder, from *client.ServiceSource) {
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

func writeStringMatch(buf *strings.Builder, prefix string, sm *client.StringMatch) {
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

func writeStringMatchMap(buf *strings.Builder, prefix string, matches map[string]client.StringMatch) {
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

func FormatURL(obj interface{}) (string, error) {
	data, ok := obj.(*Data)
	if !ok {
		return "", fmt.Errorf("invalid data")
	}
	hostBuf := strings.Builder{}
	if data.RouteSpec.Websocket {
		hostBuf.WriteString("ws://")
	} else {
		hostBuf.WriteString("http://")
	}
	hostBuf.WriteString(data.RouteSet.Name)
	if data.Stack.Name != "default" {
		hostBuf.WriteString(".")
		hostBuf.WriteString(data.Stack.Name)
	}
	hostBuf.WriteString(".")
	hostBuf.WriteString(data.Domain)
	if data.port() > 0 {
		hostBuf.WriteString(":")
		hostBuf.WriteString(strconv.FormatInt(data.port(), 10))
	}
	hostBuf.WriteString(data.path())
	return hostBuf.String(), nil
}
