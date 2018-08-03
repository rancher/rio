package stage

import (
	"fmt"

	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
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

	resource, err := lookup.Lookup(ctx.Client, app.Args()[0], client.ServiceType)
	if err != nil {
		return err
	}

	revision, err := determineRevision(app.Args()[0], resource)
	if err != nil {
		return err
	}

	args := append([]string{r.Image}, app.Args()[1:]...)
	serviceDef, err := r.ToService(args)
	if err != nil {
		return err
	}

	serviceDef.Scale = int64(r.Scale)

	newRevision := &client.ServiceRevision{}
	if err := convert.ToObj(serviceDef, newRevision); err != nil {
		return fmt.Errorf("failed to format service revision: %v", err)
	}

	service, err := ctx.Client.Service.ByID(resource.ID)
	if err != nil {
		return err
	}

	newRevision.Weight = int64(r.Weight)
	if newRevision.Scale == 0 {
		newRevision.Scale = service.Scale
	}
	if service.Revisions == nil {
		service.Revisions = map[string]client.ServiceRevision{}
	}

	service.Revisions[revision] = *newRevision

	_, err = ctx.Client.Service.Replace(service)
	if err != nil {
		return err
	}

	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	w.Add(service.ID)
	return w.Wait()
}
