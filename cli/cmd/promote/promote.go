package promote

import (
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
		resource, err := lookup.Lookup(ctx.ClientLookup, arg, client.ServiceType)
		if err != nil {
			return err
		}

		wc, err := ctx.WorkspaceClient()
		if err != nil {
			return err
		}

		parsed := lookup.ParseServiceName(arg)
		rev, err := wc.Service.ByID(resource.ID + "-" + parsed.Revision)
		if err != nil {
			return err
		}

		if p.Scale > 0 {
			rev.Scale = int64(p.Scale)
		}

		rev.Promote = true
		if _, err := wc.Service.Update(rev, rev); err != nil {
			return err
		}

		w.Add(&rev.Resource)
	}

	return w.Wait(ctx.Ctx)
}
