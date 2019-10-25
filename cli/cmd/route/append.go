package route

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/rancher/rio/cli/pkg/types"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/pretty/objectmappers"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

var (
	actions = map[string]bool{
		"to":       true,
		"mirror":   true,
		"redirect": true,
		"rewrite":  true,
	}
)

type Create struct {
	Insert bool `desc:"Insert the rule at the beginning instead of the end"`
	Add
}

type Add struct {
	Cookie          map[string]string `desc:"Match HTTP cookie (format key=value, value optional)"`
	Header          map[string]string `desc:"Match HTTP header (format key=value, value optional)"`
	FaultPercentage int               `desc:"Percentage of matching requests to fault"`
	FaultDelay      string            `desc:"Inject a delay for fault (ms|s|m|h)" default:"0s"`
	FaultHTTPCode   int               `desc:"HTTP code to send for fault injection"`
	AddHeader       []string          `desc:"Add HTTP header to request (format key=value)"`
	SetHeader       []string          `desc:"Override HTTP header to request (format key=value)"`
	RemoveHeader    []string          `desc:"Remove HTTP header to request (format key=value)"`
	RetryAttempts   int               `desc:"How many times to retry"`
	RetryTimeout    string            `desc:"Timeout per retry (ms|s|m|h)" default:"0s"`
	Timeout         string            `desc:"Timeout for all requests (ms|s|m|h)" default:"0s"`
	Method          []string          `desc:"Match HTTP method, support comma-separated values"`
}

type Action interface {
	validateServiceStack(ctx *clicontext.CLIContext, args []string) error
	buildRouteSpec(ctx *clicontext.CLIContext, args []string) (*riov1.RouteSpec, error)
	getRouteSet(ctx *clicontext.CLIContext, args []string) (*riov1.Router, bool, error)
}

func (a *Create) Run(ctx *clicontext.CLIContext) error {
	return insertRoute(ctx, a.Insert, a)
}

func insertRoute(ctx *clicontext.CLIContext, insert bool, a Action) error {
	args := ctx.CLI.Args()
	if len(args) < 3 {
		return fmt.Errorf("at least 3 arguments are required: HOST[/PATH] to|redirect|mirror|rewrite TARGET")
	}

	if err := a.validateServiceStack(ctx, args); err != nil {
		return err
	}

	routeSpec, err := a.buildRouteSpec(ctx, args)
	if err != nil {
		return err
	}

	routeSet, shouldCreate, err := a.getRouteSet(ctx, args)
	if err != nil {
		return err
	}

	if insert {
		routeSet.Spec.Routes = append([]riov1.RouteSpec{*routeSpec}, routeSet.Spec.Routes...)
	} else {
		routeSet.Spec.Routes = append(routeSet.Spec.Routes, *routeSpec)
	}

	if shouldCreate {
		return ctx.Create(routeSet)
	}

	if err := ctx.UpdateObject(routeSet); err != nil {
		return err
	}
	fmt.Printf("%s/%s\n", routeSet.Namespace, routeSet.Name)
	return nil
}

func (a *Add) validateServiceStack(ctx *clicontext.CLIContext, args []string) error {
	hostname, _ := parsePath(args[0])
	if hostname == "" {
		return fmt.Errorf("route host/path must be in the format hostname[/path], for example myservice/login")
	}

	return nil
}

func (a *Add) getRouteSet(ctx *clicontext.CLIContext, args []string) (*riov1.Router, bool, error) {
	hostname, _ := parsePath(args[0])
	namespace := ctx.GetSetNamespace()

	routeset, err := ctx.ByID(fmt.Sprintf("%s/%s", types.RouterType, hostname))
	if err != nil {
		if errors.IsNotFound(err) {
			return riov1.NewRouter(namespace, hostname, riov1.Router{}), true, nil
		}
		return nil, false, err
	}

	return routeset.Object.(*riov1.Router), false, nil
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

	if len(a.AddHeader) != 0 || len(a.SetHeader) != 0 || len(a.RemoveHeader) != 0 {
		if routeSpec.Headers == nil {
			routeSpec.Headers = &riov1.HeaderOperations{}
		}
		routeSpec.Headers.Add = stringToNameValue(a.AddHeader)
		routeSpec.Headers.Set = stringToNameValue(a.SetHeader)
		routeSpec.Headers.Remove = a.RemoveHeader
	}

	if err := a.addFault(routeSpec); err != nil {
		return nil, err
	}
	a.addMirror(routeSpec, action, destinations)
	a.addRedirect(routeSpec, action, args[2])
	a.addRewrite(routeSpec, action, args[2])
	a.addTo(routeSpec, action, destinations)

	return routeSpec, nil
}

func (a *Add) addTo(routeSpec *riov1.RouteSpec, action string, dests []riov1.WeightedDestination) {
	if action != "to" {
		return
	}

	routeSpec.To = dests
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
		App:     dests[0].App,
		Version: dests[0].Version,
		Port:    dests[0].Port,
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
		f.AbortHTTPStatus = a.FaultHTTPCode
	}

	return nil
}

func (a *Add) addMatch(ctx *clicontext.CLIContext, matchString string, routeSpec *riov1.RouteSpec) error {
	routeName, path := parsePath(matchString)
	if routeName == "" {
		return fmt.Errorf("route host/path must have a host")
	}

	addMatch := false
	match := riov1.Match{}

	if path != "" {
		addMatch = true
		if !objectmappers.IsRegexp(path) && path[0] != '/' {
			path = "/" + path
		}
		match.Path = objectmappers.ParseStringMatch(path)
	}

	if len(a.Method) > 0 {
		addMatch = true
		match.Methods = append(match.Methods, a.Method...)
	}

	if len(a.Header) > 0 {
		addMatch = true
		match.Headers = convertHeader(a.Header)
	}

	if addMatch {
		routeSpec.Match = match
	}

	return nil
}

func convertHeader(data map[string]string) []riov1.HeaderMatch {
	var result []riov1.HeaderMatch

	for name, value := range data {
		result = append(result, riov1.HeaderMatch{
			Name:  name,
			Value: objectmappers.ParseStringMatch(value),
		})
	}

	return result
}

func ParseDestinations(targets []string) ([]riov1.WeightedDestination, error) {
	var result []riov1.WeightedDestination
	for _, target := range targets {
		target, optStr := kv.Split(target, ",")
		opts := kv.SplitMap(optStr, ",")

		wd := riov1.WeightedDestination{
			Destination: riov1.Destination{
				App: target,
			},
		}

		version := opts["version"]
		wd.Version = version

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
			wd.Port = uint32(n)
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

func parsePath(str string) (string, string) {
	return kv.Split(str, "/")
}

func stringToNameValue(values []string) (r []riov1.NameValue) {
	for _, v := range values {
		nv := riov1.NameValue{}
		nv.Name, nv.Value = kv.Split(v, "=")
	}
	return r
}
