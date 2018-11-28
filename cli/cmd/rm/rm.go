package rm

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	spaceClient "github.com/rancher/rio/types/client/space/v1beta1"
)

type Rm struct {
	T_Type string `desc:"delete specific type"`
}

func (r *Rm) Run(ctx *clicontext.CLIContext) error {
	types := []string{client.ServiceType, client.StackType, spaceClient.PodType, client.ConfigType, client.VolumeType, client.ExternalServiceType}
	if len(r.T_Type) > 0 {
		types = []string{r.T_Type}
	}

	return Remove(ctx, types...)
}

func Remove(ctx *clicontext.CLIContext, types ...string) error {
	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	for _, arg := range ctx.CLI.Args() {
		resource, err := lookup.Lookup(ctx, arg, types...)
		if err != nil {
			return err
		}

		client, err := ctx.ClientLookup(resource.Type)
		if err != nil {
			return err
		}

		err = client.Delete(&resource.Resource)
		if err != nil {
			return err
		}

		w.Add(&resource.Resource)
	}

	return w.Wait(ctx.Ctx)
}
