package waiter

import (
	"fmt"
	"strings"
	"time"

	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/monitor"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
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
		Action:    waitForResources,
		Flags:     []cli.Flag{},
	}
}

func WaitFor(ctx *server.Context, resource *types.Resource) error {
	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}
	w.Add(resource)
	return w.Wait()
}

func waitForResources(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}

	ctx.CLIContext.GlobalSet("wait", "true")

	w, err := NewWaiter(ctx)
	if err != nil {
		return err
	}

	for _, r := range ctx.CLIContext.Args() {
		resource, err := lookup.Lookup(ctx.ClientLookup, r, waitTypes...)
		if err != nil {
			logrus.Info("%s does not exist")
			continue
		}
		w.Add(resource)
	}

	return w.Wait()
}

func NewWaiter(ctx *server.Context) (*Waiter, error) {
	client := ctx.Client

	waitState := ctx.CLIContext.GlobalString("wait-state")
	if waitState == "" {
		waitState = "active"
	}

	return &Waiter{
		enabled:      ctx.CLIContext.GlobalBool("wait"),
		timeout:      ctx.CLIContext.GlobalInt("wait-timeout"),
		state:        waitState,
		client:       client,
		clientLookup: ctx.ClientLookup,
	}, nil
}

type Waiter struct {
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
		fmt.Println(resource.ID)
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

func (w *Waiter) Wait() error {
	if !w.enabled {
		return nil
	}

	watching := map[ResourceID]bool{}
	w.monitor = monitor.New(w.client)
	sub := w.monitor.Subscribe()
	go func() {
		schema := w.client.Types["subscribe"]
		logrus.Fatal(w.monitor.Start(&schema))
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
		if !ok {
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
			delete(watching, resource)
		}
	}

	return nil
}
