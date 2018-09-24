package stage

import (
	"fmt"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/service"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/cli/server"
	"github.com/urfave/cli"
)

type Stage struct {
	create.Create
	Scale  int    `desc:"Number of replicas to run"`
	Weight int    `desc:"Percentage of traffic routed to staged service"`
	Image  string `desc:"Runtime image (Docker image/OCI image)"`
}

func determineRevision(name string, service *types.Resource) (string, error) {
	revision := "next"
	if name == service.ID {
		return revision, nil
	}

	parsedService := lookup.ParseServiceName(name)
	if parsedService.Revision == "latest" {
		return "", fmt.Errorf("\"latest\" is not a valid revision to stage")
	}
	if parsedService.Revision != "" {
		revision = parsedService.Revision
	}

	return revision, nil
}

func (r *Stage) Run(app *cli.Context) error {
	if len(app.Args()) == 0 {
		return fmt.Errorf("must specify the service to update")
	}

	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	service, err := service.Lookup(ctx, app.Args()[0])
	if err != nil {
		return err
	}

	revision, err := determineRevision(app.Args()[0], &service.Resource)
	if err != nil {
		return err
	}

	args := append([]string{r.Image}, app.Args()[1:]...)
	serviceDef, err := r.ToService(args)
	if err != nil {
		return err
	}

	serviceDef.ParentService = service.Name
	serviceDef.Version = revision
	serviceDef.Weight = int64(r.Weight)
	serviceDef.Scale = int64(r.Scale)
	if serviceDef.Scale == 0 {
		serviceDef.Scale = service.Scale
	}

	revService, err := ctx.Client.Service.Create(serviceDef)
	if err != nil {
		return err
	}

	w.Add(&revService.Resource)
	return w.Wait()
}
