package rm

import (
	"strings"

	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	"github.com/urfave/cli"
)

type Rm struct {
	T_Type string `desc:"delete specific type"`
}

func (r *Rm) Run(app *cli.Context) error {
	types := []string{client.ServiceType, client.StackType, client.ConfigType, client.VolumeType}
	if len(r.T_Type) > 0 {
		types = []string{r.T_Type}
	}

	return Remove(app, types...)
}

func Remove(app *cli.Context, types ...string) error {
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
		resource, err := lookup.Lookup(ctx.Client, arg, types...)
		if err != nil {
			return err
		}

		if arg != resource.ID && resource.Type == client.ServiceType && strings.Contains(arg, ":") {
			parsed := lookup.ParseServiceName(arg)
			service, err := ctx.Client.Service.ByID(resource.ID)
			if err != nil {
				return err
			}
			if _, ok := service.Revisions[parsed.Revision]; ok {
				delete(service.Revisions, parsed.Revision)
				if _, err := ctx.Client.Service.Replace(service); err != nil {
					return err
				}
			}

			w.Add(resource.ID)
			continue
		}

		err = ctx.Client.Delete(resource)
		if err != nil {
			return err
		}
		w.Add(resource.ID)
	}

	return w.Wait()
}
