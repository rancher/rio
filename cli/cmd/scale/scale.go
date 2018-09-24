package scale

import (
	"fmt"
	"strconv"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Scale struct {
}

func (s *Scale) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	waiter, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	for _, arg := range app.Args() {
		name, scaleStr := kv.Split(arg, "=")
		scale, err := strconv.Atoi(scaleStr)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %v", arg, err)
		}

		resource, err := lookup.Lookup(ctx.ClientLookup, name, client.ServiceType)
		if err != nil {
			return err
		}

		service, err := ctx.Client.Service.ByID(resource.ID)
		if err != nil {
			return err
		}

		parsedService := lookup.ParseServiceName(name)
		if parsedService.Revision != "" && parsedService.Revision != service.Version {
			service, err = ctx.Client.Service.ByID(resource.ID + "-" + service.Version)
			if err != nil {
				return err
			}
		}

		err = ctx.Client.Update(client.ServiceType, resource, map[string]interface{}{
			"scale": scale,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to update scale on %s: %v", name, err)
		}

		waiter.Add(&service.Resource)
	}

	return waiter.Wait()
}
