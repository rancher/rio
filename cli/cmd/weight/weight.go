package weight

import (
	"errors"
	"fmt"
	"strconv"

	"strings"

	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Weight struct {
}

func (w *Weight) Run(app *cli.Context) error {
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
		scaleStr = strings.TrimSuffix(scaleStr, "%")

		if scaleStr == "" {
			return fmt.Errorf("weight params must be in the format of SERVICE=PERCENTAGE, for example: mystack/myservice=10%")
		}
		scale, err := strconv.Atoi(scaleStr)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %v", arg, err)
		}

		resource, err := lookup.Lookup(ctx.Client, name, client.ServiceType)
		if err != nil {
			return err
		}

		service, err := ctx.Client.Service.ByID(resource.ID)
		if err != nil {
			return err
		}

		parsedService := lookup.ParseServiceName(name)
		if rev, ok := service.Revisions[parsedService.Revision]; ok {
			rev.Weight = int64(scale)
			service.Revisions[parsedService.Revision] = rev
			_, err = ctx.Client.Service.Update(service, service)
		} else {
			return errors.New("weight can only be added to staged services")
		}

		waiter.Add(service.ID)
	}

	return waiter.Wait()
}
