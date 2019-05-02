package route

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/pretty/objectmappers"
	"github.com/rancher/wrangler/pkg/kv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	actions = map[string]bool{
		"to":       true,
		"mirror":   true,
		"redirect": true,
		"rewrite":  true,
	}
)

type Append struct {
	Add
}

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
}

type RouteAction interface {
	validateServiceStack(ctx *clicontext.CLIContext, args []string) error
	buildRouteSpec(ctx *clicontext.CLIContext, args []string) (*riov1.RouteSpec, error)
	getRouteSet(ctx *clicontext.CLIContext, args []string) (*riov1.Router, error)
}

func (a *Append) Run(ctx *clicontext.CLIContext) error {
	return insertRoute(ctx, false, a)
}

func insertRoute(ctx *clicontext.CLIContext, insert bool, a RouteAction) error {
	args := ctx.CLI.Args()
	if len(args) < 3 {
		return fmt.Errorf("at least 3 arguements are required: HOST[/PATH] to|redirect|mirror TARGET")
	}

	if err := a.validateServiceStack(ctx, args); err != nil {
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

	if insert {
		routeSet.Spec.Routes = append([]riov1.RouteSpec{*routeSpec}, routeSet.Spec.Routes...)
	} else {
		routeSet.Spec.Routes = append(routeSet.Spec.Routes, *routeSpec)
	}

	if routeSet.Name == "" {
		return ctx.Create(routeSet)
	}

	return ctx.UpdateObject(routeSet)
}

func (a *Add) validateServiceStack(ctx *clicontext.CLIContext, args []string) error {
	_, service, stack, _, _ := parseMatch(args[0])
	if service == "" {
		return fmt.Errorf("route host/path must be in the format service.stack[/path], for example myservice.mystack/login")
	}
	if stack == "" {
		stack = ctx.GetDefaultStackName()
	}

	return nil
}

func (a *Add) getRouteSet(ctx *clicontext.CLIContext, args []string) (*riov1.Router, error) {
	_, service, stackName, _, _ := parseMatch(args[0])

	_, stackName, name, err := stack.NamespaceAndName(ctx, stackName+"/"+service)
	if err != nil {
		return nil, err
	}

	routeSet, err := lookupRoute(ctx, stackName, name)
	if err != nil {
		return nil, err
	}

	if routeSet != nil {
		return routeSet, nil
	}

	return riov1.NewRouter(stackName, name, riov1.Router{}), nil
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

func (a *Add) buildRouteSpec(ctx *clicontext.CLIContext, args []string) (*riov1.RouteSpec, error) {
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

	routeSpec := &riov1.RouteSpec{}
	if err := a.addMatch(ctx, args[0], routeSpec); err != nil {
		return nil, err
	}

	routeSpec.AddHeaders = a.AddHeader
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

func (a *Add) addTo(routeSpec *riov1.RouteSpec, action string, dests []riov1.WeightedDestination) {
	if action != "to" {
		return
	}

	routeSpec.To = dests
}

func (a *Add) addTimeout(routeSpec *riov1.RouteSpec) error {
	n, err := objectmappers.ParseDurationUnit(a.Timeout, "timeout", time.Millisecond)
	if err != nil {
		return err
	}

	routeSpec.TimeoutMillis = &n
	return nil
}

func (a *Add) addRedirect(routeSpec *riov1.RouteSpec, action string, dest string) {
	if action != "redirect" {
		return
	}

	host, path := kv.Split(dest, "/")
	if path != "" {
		path = "/" + path
	}
	routeSpec.Redirect = &riov1.Redirect{
		Path: path,
		Host: host,
	}
}

func (a *Add) addRewrite(routeSpec *riov1.RouteSpec, action string, dest string) {
	if action != "rewrite" {
		return
	}

	host, path := kv.Split(dest, "/")
	if path != "" {
		path = "/" + path
	}
	routeSpec.Rewrite = &riov1.Rewrite{
		Path: path,
		Host: host,
	}
}

func (a *Add) addMirror(routeSpec *riov1.RouteSpec, action string, dests []riov1.WeightedDestination) {
	if action != "mirror" {
		return
	}

	routeSpec.Mirror = &riov1.Destination{
		Stack:    dests[0].Stack,
		Service:  dests[0].Service,
		Revision: dests[0].Revision,
		Port:     dests[0].Port,
	}
}

func (a *Add) addFault(routeSpec *riov1.RouteSpec) error {
	if a.FaultPercentage <= 0 {
		return nil
	}

	f := &riov1.Fault{
		Percentage: a.FaultPercentage,
	}

	if a.FaultDelay != "0s" && a.FaultDelay != "" {
		d, err := objectmappers.ParseDurationUnit(a.FaultDelay, "fault delay", time.Millisecond)
		if err != nil {
			return err
		}
		f.DelayMillis = d
		return nil
	}

	if a.FaultHTTPCode != 0 {
		f.Abort = riov1.Abort{
			HTTPStatus: a.FaultHTTPCode,
		}
		return nil
	}

	if a.FaultHTTP2Error != "" {
		f.Abort = riov1.Abort{
			HTTP2Status: a.FaultHTTP2Error,
		}
		return nil
	}

	if a.FaultGRPCError != "" {
		f.Abort = riov1.Abort{
			GRPCStatus: a.FaultGRPCError,
		}
		return nil
	}

	return nil
}

func (a *Add) addMatch(ctx *clicontext.CLIContext, matchString string, routeSpec *riov1.RouteSpec) error {
	scheme, service, _, path, port := parseMatch(matchString)
	if service == "" {
		return fmt.Errorf("route host/path must have a host in the format of SERVICE.STACK, for example myservice.mystack")
	}

	addMatch := false
	match := riov1.Match{}

	if scheme != "" {
		addMatch = true
		match.Scheme = objectmappers.ParseStringMatch(scheme)
	}

	if path != "" {
		addMatch = true
		if !objectmappers.IsRegexp(path) && path[0] != '/' {
			path = "/" + path
		}
		match.Path = objectmappers.ParseStringMatch(path)
	}

	if a.Method != "" {
		addMatch = true
		match.Method = objectmappers.ParseStringMatch(a.Method)
	}

	if a.From != "" {
		addMatch = true
		wds, err := ParseDestinations([]string{a.From})
		if err != nil {
			return fmt.Errorf("invalid format for --from [%s]: %v", a.From, err)
		}
		match.From = &riov1.ServiceSource{
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
		match.Port = &[]int{int(n)}[0]
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

func stringMapToStringMatchMap(data map[string]string) map[string]riov1.StringMatch {
	result := map[string]riov1.StringMatch{}

	for k, v := range data {
		result[k] = *objectmappers.ParseStringMatch(v)
	}

	return result
}

func ParseDestinations(targets []string) ([]riov1.WeightedDestination, error) {
	var result []riov1.WeightedDestination
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

		wd := riov1.WeightedDestination{
			Destination: riov1.Destination{
				Stack:    stack,
				Service:  service,
				Revision: revision,
			},
		}

		weight := opts["weight"]
		if weight != "" {
			n, err := strconv.ParseInt(strings.TrimSuffix(weight, "%"), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid weight format [%s]", weight)
			}
			wd.Weight = int(n)
		}

		port := opts["port"]
		if port != "" {
			n, err := strconv.ParseInt(port, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid port format [%s]", port)
			}
			wd.Port = &[]uint32{uint32(n)}[0]
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

func lookupRoute(ctx *clicontext.CLIContext, stackName, name string) (*riov1.Router, error) {
	route, err := ctx.Rio.Routers(stackName).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return route, nil
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
