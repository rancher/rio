package weight

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/norman/types/values"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

type Weight struct {
}

func (w *Weight) Run(ctx *clicontext.CLIContext) error {
	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	waiter, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	for _, arg := range ctx.CLI.Args() {
		name, scaleStr := kv.Split(arg, "=")
		scaleStr = strings.TrimSuffix(scaleStr, "%")

		if scaleStr == "" {
			return fmt.Errorf("weight params must be in the format of SERVICE=PERCENTAGE, for example: mystack/myservice=10%%")
		}
		scale, err := strconv.Atoi(scaleStr)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %v", arg, err)
		}

		service, err := lookup.Lookup(ctx, name, client.ServiceType)
		if err != nil {
			return err
		}

		data := map[string]interface{}{}
		values.PutValue(data, int64(scale),
			client.ServiceFieldWeight)

		_, err = wc.Service.Update(&client.Service{Resource: service.Resource}, data)
		if err != nil {
			return err
		}

		waiter.Add(&service.Resource)
	}

	return waiter.Wait(ctx.Ctx)
}
