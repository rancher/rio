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
		resource, err := lookup.Lookup(ctx.Client, arg, client.ServiceType)
		if err != nil {
			return err
		}

		service, err := ctx.Client.Service.ByID(resource.ID)
		if err != nil {
			return err
		}

		parsed := lookup.ParseServiceName(arg)
		if rev, ok := service.Revisions[parsed.Revision]; ok {
			if p.Scale > 0 {
				service.Scale = int64(p.Scale)
			}

			rev.Promote = true
			service.Revisions[parsed.Revision] = rev
			_, err := ctx.Client.Service.Update(service, service)
			if err != nil {
				return err
			}
			w.Add(resource.ID)
		}
	}

	return w.Wait()
}
