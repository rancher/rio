package waiter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/monitor"
	"github.com/rancher/rio/types/client/rio/v1"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	waitTypes = []string{"service", "stack"}
)

func WaitCommand() cli.Command {
	return cli.Command{
		Name:      "wait",
		Usage:     "Wait for resources " + strings.Join(waitTypes, ", "),
		ArgsUsage: "[ID NAME...]",
		Action:    clicontext.Wrap(waitForResources),
		Flags:     []cli.Flag{},
	}
}

func waitForResources(ctx *clicontext.CLIContext) error {
	ctx.Config.Wait = true

	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}

	timeout := ctx.Config.WaitTimeout
	end := time.Now().Add(time.Duration(timeout) * time.Second)

	for _, r := range ctx.CLI.Args() {
		for {
			resource, err := lookup.Lookup(ctx, r, waitTypes...)
			if err != nil {
				if time.Now().After(end) {
					return fmt.Errorf("timeout waiting for %s to exist", r)
				}
				time.Sleep(time.Second)
				continue
			}
			w.Add(&resource.Resource)
			break
		}
	}

	return w.Wait(ctx.Ctx)
}

func WaitFor(ctx *clicontext.CLIContext, objs ...*types.Resource) error {
	return waitFor(ctx, false, objs...)
}

func EnsureActive(ctx *clicontext.CLIContext, objs ...*types.Resource) error {
	return waitFor(ctx, true, objs...)
}

func waitFor(ctx *clicontext.CLIContext, enable bool, objs ...*types.Resource) error {
	clientCtx, cancel := context.WithCancel(ctx.Ctx)
	defer cancel()

	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}
	if enable {
		w.quiet = true
		w.enabled = true
	}
	w.state = "active"

	w.Add(objs...)
	return w.Wait(clientCtx)
}

func NewWaiter(ctx *clicontext.CLIContext) (*Waiter, error) {
	wc, err := ctx.ProjectClient()
	if err != nil {
		return nil, err
	}

	waitState := ctx.Config.WaitState
	if waitState == "" {
		waitState = "active"
	}

	return &Waiter{
		enabled:      ctx.Config.Wait,
		timeout:      ctx.Config.WaitTimeout,
		state:        waitState,
		client:       wc,
		clientLookup: ctx,
	}, nil
}

type Waiter struct {
	quiet        bool
	enabled      bool
	timeout      int
	state        string
	resources    []*types.Resource
	client       *client.Client
	clientLookup lookup.ClientLookup
	monitor      *monitor.Monitor
}

type ResourceID string

func NewResourceID(resourceType, id string) ResourceID {
	return ResourceID(fmt.Sprintf("%s:%s", resourceType, id))
}

func (r ResourceID) ID() string {
	str := string(r)
	return str[strings.Index(str, ":")+1:]
}

func (r ResourceID) Type() string {
	str := string(r)
	return str[:strings.Index(str, ":")]
}

func (w *Waiter) Add(resources ...*types.Resource) *Waiter {
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		if !w.quiet {
			fmt.Println(resource.ID)
		}
		w.resources = append(w.resources, resource)
	}
	return w
}

func (w *Waiter) done(resourceType, id string) (bool, error) {
	data := map[string]interface{}{}
	ok, err := w.monitor.Get(resourceType, id, &data)
	if err != nil {
		return ok, err
	}

	if ok {
		return w.checkDone(resourceType, id, data)
	}

	if err := w.client.ByID(resourceType, id, &data); err != nil {
		if clientbase.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}

	return w.checkDone(resourceType, id, data)
}

func (w *Waiter) checkDone(resourceType, id string, data map[string]interface{}) (bool, error) {
	transitioning := fmt.Sprint(data["transitioning"])
	logrus.Debugf("%s:%s transitioning=%s state=%v, healthState=%v waitState=%s", resourceType, id, transitioning,
		data["state"], data["healthState"], w.state)

	switch transitioning {
	case "yes":
		return false, nil
	case "error":
		return false, fmt.Errorf("%s:%s failed: %s", resourceType, id, data["transitioningMessage"])
	}

	if w.state == "" {
		return true, nil
	}

	return data["state"] == w.state || data["healthState"] == w.state, nil
}

func (w *Waiter) Wait(ctx context.Context) error {
	if !w.enabled {
		return nil
	}

	watching := map[ResourceID]bool{}
	w.monitor = monitor.New(w.client)
	sub := w.monitor.Subscribe()
	go func() {
		schema := w.client.Types["subscribe"]
		for {
			err := w.monitor.Start(ctx, &schema)
			select {
			case <-ctx.Done():
				return
			default:
				logrus.Errorf("subscription failed, will retry: %v", err)
			}
		}
	}()

	for _, resource := range w.resources {
		r, err := lookup.Lookup(w.clientLookup, resource.ID, resource.Type)
		if err != nil {
			return err
		}

		ok, err := w.done(r.Type, r.ID)
		if err != nil {
			return err
		}
		if ok {
			logrus.Debugf("resource %s:%s already done", r.Type, r.ID)
		} else {
			watching[NewResourceID(r.Type, r.ID)] = true
		}
	}

	timeout := time.After(time.Duration(w.timeout) * time.Second)
	every := time.Tick(2 * time.Second)
	for len(watching) > 0 {
		var event *monitor.Event
		select {
		case event = <-sub.C:
		case <-timeout:
			return fmt.Errorf("timeout")
		case <-every:
			for resource := range watching {
				ok, err := w.done(resource.Type(), resource.ID())
				if err != nil {
					return err
				}
				if ok {
					logrus.Debugf("resource %s:%s done from polling", resource.Type(), resource.ID())
					delete(watching, resource)
				}
			}
			continue
		}

		resourceType := event.ResourceType()
		resourceID := event.ResourceID()

		resource := NewResourceID(resourceType, resourceID)
		if !watching[resource] {
			continue
		}

		done, err := w.done(resourceType, resourceID)
		if err != nil {
			return err
		}

		if done {
			logrus.Debugf("resource %s:%s done from event", resource.Type(), resource.ID())
			delete(watching, resource)
		}
	}

	return nil
}
