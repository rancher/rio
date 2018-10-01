package scale

import (
	"fmt"
	"strconv"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

type Scale struct {
}

func (s *Scale) Run(ctx *clicontext.CLIContext) error {
	waiter, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	for _, arg := range ctx.CLI.Args() {
		name, scaleStr := kv.Split(arg, "=")
		scale, err := strconv.Atoi(scaleStr)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %v", arg, err)
		}

		resource, err := lookup.Lookup(ctx.ClientLookup, name, client.ServiceType)
		if err != nil {
			return err
		}

		service, err := wc.Service.ByID(resource.ID)
		if err != nil {
			return err
		}

		parsedService := lookup.ParseServiceName(name)
		if parsedService.Revision != "" && parsedService.Revision != service.Version {
			service, err = wc.Service.ByID(resource.ID + "-" + service.Version)
			if err != nil {
				return err
			}
		}

		err = wc.Update(client.ServiceType, resource, map[string]interface{}{
			"scale": scale,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to update scale on %s: %v", name, err)
		}

		waiter.Add(&service.Resource)
	}

	return waiter.Wait(ctx.Ctx)
}
