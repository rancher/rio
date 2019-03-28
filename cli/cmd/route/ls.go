package route

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/pkg/namespace"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Data struct {
	ID        string
	Stack     *riov1.Stack
	RouteSet  riov1.RouteSet
	RouteSpec riov1.RouteSpec
	Match     *riov1.Match
	Domain    string
}

func (d *Data) port() int {
	if d.Match == nil || d.Match.Port == nil {
		return 0
	}
	return *d.Match.Port
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

func stringMatchToString(m *riov1.StringMatch) string {
	if m == nil {
		return ""
	}
	return v1.StringMatch{
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

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	project, err := ctx.Project()
	if err != nil {
		return err
	}

	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}

	domain, err := cluster.Domain()
	if err != nil {
		return err
	}

	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	routeSets, err := client.Rio.RouteSets("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	writer := table.NewWriter([][]string{
		{"NAME", "{{stackScopedName .Stack.Name .RouteSet.Name}}"},
		{"URL", "{{ . | formatURL }}"},
		{"OPTS", "{{ . | formatOpts }}"},
		{"ACTION", "{{ . | formatAction }}"},
		{"TARGET", "{{ . | formatTarget }}"},
	}, ctx, os.Stdout)
	defer writer.Close()

	writer.AddFormatFunc("formatURL", FormatURL())
	writer.AddFormatFunc("formatOpts", FormatOpts)
	writer.AddFormatFunc("formatAction", FormatAction)
	writer.AddFormatFunc("formatTarget", FormatTarget)
	writer.AddFormatFunc("stackScopedName", table.FormatStackScopedName(cluster))

	stackByID, err := util.StacksByID(client, project.Project.Name)
	if err != nil {
		return err
	}

	for _, routeSet := range routeSets.Items {
		stack := stackByID[routeSet.Spec.StackName]
		if stack == nil {
			continue
		}
		for j, routeSpec := range routeSet.Spec.Routes {
			if len(routeSpec.Matches) == 0 {
				writer.Write(&Data{
					ID:        routeSet.Name,
					RouteSet:  routeSet,
					RouteSpec: routeSet.Spec.Routes[j],
					Stack:     stackByID[routeSet.Spec.StackName],
					Domain:    domain,
				})
				continue
			}

			for k := range routeSpec.Matches {
				writer.Write(&Data{
					RouteSet:  routeSet,
					RouteSpec: routeSet.Spec.Routes[j],
					Match:     &routeSet.Spec.Routes[j].Matches[k],
					Stack:     stackByID[routeSet.Spec.StackName],
					Domain:    domain,
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

func FormatTarget(obj interface{}) (string, error) {
	buf := &strings.Builder{}
	data, ok := obj.(*Data)
	if !ok {
		return "", fmt.Errorf("invalid data")
	}

	target := targetType(data)

	if target == "to" {
		for _, to := range data.RouteSpec.To {
			writeDest(buf, data.Stack.Name, to.Stack, to.Service, to.Revision, int(*to.Port), to.Weight)
		}
	} else if target == "redirect" && data.RouteSpec.Redirect != nil {
		writeHostPath(buf, data.RouteSpec.Redirect.Host, data.RouteSpec.Redirect.Path)
	} else if target == "mirror" && data.RouteSpec.Mirror != nil && data.RouteSpec.Mirror.Port != nil {
		writeDest(buf, data.Stack.Name, data.RouteSpec.Mirror.Stack, data.RouteSpec.Mirror.Service,
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

func addFrom(buf *strings.Builder, from *riov1.ServiceSource) {
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

func writeStringMatch(buf *strings.Builder, prefix string, sm *riov1.StringMatch) {
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

func writeStringMatchMap(buf *strings.Builder, prefix string, matches map[string]riov1.StringMatch) {
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
		data, ok := obj.(*Data)
		if !ok {
			return "", fmt.Errorf("invalid data")
		}
		hostBuf := strings.Builder{}
		hostBuf.WriteString("https://")
		name := namespace.HashIfNeed(data.RouteSet.Name, data.Stack.Name, data.Stack.Namespace)
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
