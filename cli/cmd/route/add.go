package route

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"sort"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

var (
	actions = map[string]bool{
		"to":       true,
		"mirror":   true,
		"redirect": true,
		"rewrite":  true,
	}
)

type Add struct {
	Cookie          map[string]string `desc:"Match HTTP cookie (format key=value, value optional)"`
	Header          map[string]string `desc:"Match HTTP header (format key=value, value optional)"`
	FaultPercentage int               `desc:"Percentage of matching requests to fault"`
	FaultDelay      string            `desc:"Inject a delay for fault (ms|s|m|h)" default:"0s"`
	FaultHTTPCode   int               `desc:"HTTP code to send for fault injection"`
	FaultHTTP2Error string            `desc:"HTTP2 error to send for fault injection"`
	FaultGRPCError  string            `desc:"gRPC error to send for fault injection"`
	AddHeader       []string          `desc:"Add HTTP header to request (format key=value)"`
	RetryAttempts   int               `desc:"How many times to retry"`
	RetryTimeout    string            `desc:"Timeout per retry (ms|s|m|h)" default:"0s"`
	Timeout         string            `desc:"Timeout for all requests (ms|s|m|h)" default:"0s"`
	Method          string            `desc:"Match HTTP method"`
	From            string            `desc:"Match traffic from specific service"`
	Websocket       bool              `desc:"Websocket request"`
}

func (a *Add) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	args := app.Args()
	if len(args) < 3 {
		return fmt.Errorf("at least 3 arguements are required: HOST[/PATH] to|redirect|mirror TARGET")
	}

	if err := a.validateServiceStack(args); err != nil {
		return err
	}

	routeSpec, err := a.buildRouteSpec(ctx, args)
	if err != nil {
		return err
	}

	routeSet, err := a.getRouteSet(ctx, args)
	if err != nil {
		return err
	}

	routeSet.Routes = append(routeSet.Routes, *routeSpec)

	if routeSet.ID == "" {
		routeSet, err = ctx.Client.RouteSet.Create(routeSet)
	} else {
		routeSet, err = ctx.Client.RouteSet.Replace(routeSet)
	}

	if err == nil {
		fmt.Println(routeSet.ID)
	}
	return err
}

func (a *Add) validateServiceStack(args []string) error {
	_, service, stack, _, _ := parseMatch(args[0])
	if service == "" {
		return fmt.Errorf("route host/path must be in the format service.stack[/path], for example myservice.mystack/login")
	}
	if stack == "" {
		stack = "default"
	}

	return nil
}

func (a *Add) getRouteSet(ctx *server.Context, args []string) (*client.RouteSet, error) {
	_, service, stack, _, _ := parseMatch(args[0])

	spaceID, stackID, name, err := ctx.ResolveSpaceStackName(stack + "/" + service)
	if err != nil {
		return nil, err
	}

	routeSet, err := lookupRoute(ctx, spaceID, stackID, name)
	if err != nil {
		return nil, err
	}

	if routeSet != nil {
		return routeSet, nil
	}

	return &client.RouteSet{
		Name:    name,
		SpaceID: spaceID,
		StackID: stackID,
	}, nil
}

func actionsString(many bool) string {
	var s []string
	for k := range actions {
		if many || k != "to" {
			s = append(s, k)
		}
	}
	sort.Strings(s)
	return strings.Join(s, ", ")
}

func (a *Add) buildRouteSpec(ctx *server.Context, args []string) (*client.RouteSpec, error) {
	action, err := parseAction(args[1])
	if err != nil {
		return nil, err
	}

	if action != "to" && len(args) != 3 {
		return nil, fmt.Errorf("for %s actions only one target is allowed, found %d", actionsString(false), len(args)-2)
	}

	destinations, err := ParseDestinations(args[2:])
	if err != nil {
		return nil, err
	}

	routeSpec := &client.RouteSpec{}
	if err := a.addMatch(ctx, args[0], routeSpec); err != nil {
		return nil, err
	}

	routeSpec.AddHeaders = a.AddHeader
	routeSpec.Websocket = a.Websocket
	if err := a.addFault(routeSpec); err != nil {
		return nil, err
	}
	a.addMirror(routeSpec, action, destinations)
	a.addRedirect(routeSpec, action, args[2])
	a.addRewrite(routeSpec, action, args[2])
	a.addTo(routeSpec, action, destinations)
	if err := a.addTimeout(routeSpec); err != nil {
		return nil, err
	}

	return routeSpec, nil
}

func (a *Add) addTo(routeSpec *client.RouteSpec, action string, dests []client.WeightedDestination) {
	if action != "to" {
		return
	}

	routeSpec.To = dests
}

func (a *Add) addTimeout(routeSpec *client.RouteSpec) error {
	n, err := create.ParseDurationUnit(a.Timeout, "timeout", time.Millisecond)
	if err != nil {
		return err
	}

	routeSpec.TimeoutMillis = n
	return nil
}

func (a *Add) addRedirect(routeSpec *client.RouteSpec, action string, dest string) {
	if action != "redirect" {
		return
	}

	host, path := kv.Split(dest, "/")
	if path != "" {
		path = "/" + path
	}
	routeSpec.Redirect = &client.Redirect{
		Path: path,
		Host: host,
	}
}

func (a *Add) addRewrite(routeSpec *client.RouteSpec, action string, dest string) {
	if action != "rewrite" {
		return
	}

	host, path := kv.Split(dest, "/")
	if path != "" {
		path = "/" + path
	}
	routeSpec.Rewrite = &client.Rewrite{
		Path: path,
		Host: host,
	}
}

func (a *Add) addMirror(routeSpec *client.RouteSpec, action string, dests []client.WeightedDestination) {
	if action != "mirror" {
		return
	}

	routeSpec.Mirror = &client.Destination{
		Stack:    dests[0].Stack,
		Service:  dests[0].Service,
		Revision: dests[0].Revision,
		Port:     dests[0].Port,
	}
}

func (a *Add) addFault(routeSpec *client.RouteSpec) error {
	if a.FaultPercentage <= 0 {
		return nil
	}

	f := &client.Fault{
		Percentage: int64(a.FaultPercentage),
	}

	if a.FaultDelay != "0s" && a.FaultDelay != "" {
		d, err := create.ParseDurationUnit(a.FaultDelay, "fault delay", time.Millisecond)
		if err != nil {
			return err
		}
		f.DelayMillis = d
		return nil
	}

	if a.FaultHTTPCode != 0 {
		f.Abort = &client.Abort{
			HTTPStatus: int64(a.FaultHTTPCode),
		}
		return nil
	}

	if a.FaultHTTP2Error != "" {
		f.Abort = &client.Abort{
			HTTP2Status: a.FaultHTTP2Error,
		}
		return nil
	}

	if a.FaultGRPCError != "" {
		f.Abort = &client.Abort{
			GRPCStatus: a.FaultGRPCError,
		}
		return nil
	}

	return nil
}

func (a *Add) addMatch(ctx *server.Context, matchString string, routeSpec *client.RouteSpec) error {
	scheme, service, _, path, port := parseMatch(matchString)
	if service == "" {
		return fmt.Errorf("route host/path must have a host in the format of SERVICE.STACK, for example myservice.mystack")
	}

	addMatch := false
	match := client.Match{}

	if scheme != "" {
		addMatch = true
		match.Scheme = create.ParseStringMatch(scheme)
	}

	if path != "" {
		addMatch = true
		if !create.IsRegexp(path) && path[0] != '/' {
			path = "/" + path
		}
		match.Path = create.ParseStringMatch(path)
	}

	if a.Method != "" {
		addMatch = true
		match.Method = create.ParseStringMatch(a.Method)
	}

	if a.From != "" {
		addMatch = true
		wds, err := ParseDestinations([]string{a.From})
		if err != nil {
			return fmt.Errorf("invalid format for --from [%s]: %v", a.From, err)
		}
		match.From = &client.ServiceSource{
			Stack:    wds[0].Stack,
			Service:  wds[0].Service,
			Revision: wds[0].Revision,
		}
	}

	if port != "" {
		addMatch = true
		n, err := strconv.ParseInt(port, 10, 0)
		if err != nil {
			return fmt.Errorf("invalid port number in host/path [%s]: %s", matchString, port)
		}
		match.Port = n
	}

	if len(a.Header) > 0 {
		addMatch = true
		match.Headers = stringMapToStringMatchMap(a.Header)
	}

	if len(a.Cookie) > 0 {
		addMatch = true
		match.Cookies = stringMapToStringMatchMap(a.Cookie)
	}

	if addMatch {
		routeSpec.Matches = append(routeSpec.Matches, match)
	}

	return nil
}

func stringMapToStringMatchMap(data map[string]string) map[string]client.StringMatch {
	result := map[string]client.StringMatch{}

	for k, v := range data {
		result[k] = *create.ParseStringMatch(v)
	}

	return result
}

func ParseDestinations(targets []string) ([]client.WeightedDestination, error) {
	var result []client.WeightedDestination
	for _, target := range targets {
		var (
			stack    string
			service  string
			revision string
		)

		target, optStr := kv.Split(target, ",")
		opts := kv.SplitMap(optStr, ",")

		parts := strings.SplitN(target, "/", 2)
		if len(parts) == 2 {
			stack = parts[0]
			service = parts[1]
		} else {
			service = parts[0]
		}

		service, revision = kv.Split(service, ":")

		wd := client.WeightedDestination{
			Stack:    stack,
			Service:  service,
			Revision: revision,
		}

		weight := opts["weight"]
		if weight != "" {
			n, err := strconv.ParseInt(strings.TrimSuffix(weight, "%"), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid weight format [%s]", weight)
			}
			wd.Weight = n
		}

		port := opts["port"]
		if port != "" {
			n, err := strconv.ParseInt(port, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid port format [%s]", port)
			}
			wd.Port = n
		}

		result = append(result, wd)
	}

	return result, nil
}

func parseAction(action string) (string, error) {
	if !actions[action] {
		return "", fmt.Errorf("invalid action %s, action must be %s", action, actionsString(true))
	}
	return action, nil
}

func lookupRoute(ctx *server.Context, spaceID, stackID, name string) (*client.RouteSet, error) {
	resp, err := ctx.Client.RouteSet.List(&types.ListOpts{
		Filters: map[string]interface{}{
			client.RouteSetFieldName:    name,
			client.RouteSetFieldSpaceID: spaceID,
			client.RouteSetFieldStackID: stackID,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) != 1 {
		return nil, nil
	}

	return &resp.Data[0], nil
}

func parseMatch(str string) (scheme string, service string, stack string, path string, port string) {
	parts := strings.SplitN(str, "://", 2)
	if len(parts) == 2 {
		scheme = parts[0]
		str = parts[1]
	}

	str, path = kv.Split(str, "/")
	service, stack = kv.Split(str, ".")
	stack, port = kv.Split(stack, ":")
	return
}
