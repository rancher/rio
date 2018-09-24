package promote

import (
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Promote struct {
	Scale int `desc:"scale of service after promotion"`
}

func (p *Promote) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	for _, arg := range app.Args() {
		resource, err := lookup.Lookup(ctx.ClientLookup, arg, client.ServiceType)
		if err != nil {
			return err
		}

		parsed := lookup.ParseServiceName(arg)
		rev, err := ctx.Client.Service.ByID(resource.ID + "-" + parsed.Revision)
		if err != nil {
			return err
		}

		if p.Scale > 0 {
			rev.Scale = int64(p.Scale)
		}

		rev.Promote = true
		if _, err := ctx.Client.Service.Update(rev, rev); err != nil {
			return err
		}

		w.Add(&rev.Resource)
	}

	return w.Wait()
}
