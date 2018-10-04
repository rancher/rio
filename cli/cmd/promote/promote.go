package promote

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

type Promote struct {
	Scale int `desc:"scale of service after promotion"`
}

func (p *Promote) Run(ctx *clicontext.CLIContext) error {
	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	for _, arg := range ctx.CLI.Args() {
		resource, err := lookup.Lookup(ctx, arg, client.ServiceType)
		if err != nil {
			return err
		}

		wc, err := ctx.WorkspaceClient()
		if err != nil {
			return err
		}

		service, err := wc.Service.ByID(resource.ID)
		if err != nil {
			return err
		}

		if service.ParentService == "" {
			return fmt.Errorf("can not promote the base version")
		}

		updates := &client.Service{
			Promote: true,
		}
		if p.Scale > 0 {
			updates.Scale = int64(p.Scale)
		}

		service, err = wc.Service.Update(&client.Service{Resource: resource.Resource}, updates)
		if err != nil {
			return err
		}

		w.Add(&service.Resource)
	}

	return w.Wait(ctx.Ctx)
}
