package scale

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/waiter"
	client "github.com/rancher/rio/types/client/rio/v1"
)

type Scale struct {
}

func (s *Scale) Run(ctx *clicontext.CLIContext) error {
	waiter, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	wc, err := ctx.ProjectClient()
	if err != nil {
		return err
	}

	for _, arg := range ctx.CLI.Args() {
		name, scaleStr := kv.Split(arg, "=")
		update := map[string]interface{}{}
		resource, err := lookup.Lookup(ctx, name, client.ServiceType)
		if err != nil {
			return err
		}
		service, err := wc.Service.ByID(resource.ID)
		if err != nil {
			return err
		}
		if strings.ContainsRune(scaleStr, '-') {
			min, max := kv.Split(scaleStr, "-")
			minScale, _ := strconv.Atoi(min)
			maxScale, _ := strconv.Atoi(max)
			concurrency := 10
			if service.AutoScale != nil {
				concurrency = int(service.AutoScale.Concurrency)
			}
			update["autoScale"] = map[string]interface{}{
				"minScale":    minScale,
				"maxScale":    maxScale,
				"concurrency": concurrency,
			}
		} else {
			scale, err := strconv.Atoi(scaleStr)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %v", arg, err)
			}
			update["scale"] = scale
		}

		err = wc.Update(client.ServiceType, &resource.Resource, update, nil)
		if err != nil {
			return fmt.Errorf("failed to update scale on %s: %v", name, err)
		}

		waiter.Add(&resource.Resource)
	}

	return waiter.Wait(ctx.Ctx)
}
