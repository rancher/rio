package stage

import (
	"fmt"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

type Stage struct {
	create.Create
	Scale  int    `desc:"Number of replicas to run"`
	Weight int    `desc:"Percentage of traffic routed to staged service"`
	Image  string `desc:"Runtime image (Docker image/OCI image)"`
}

func determineRevision(workspace *clientcfg.Workspace, name string, service *types.Resource) (string, error) {
	revision := "next"
	if name == service.ID {
		return revision, nil
	}

	parsedService := lookup.ParseStackScoped(workspace, name)
	if parsedService.Version == settings.DefaultServiceVersion {
		return "", fmt.Errorf("\"%s\" is not a valid revision to stage", settings.DefaultServiceVersion)
	}
	if parsedService.Version != "" {
		revision = parsedService.Version
	}

	return revision, nil
}

func stripRevision(workspace *clientcfg.Workspace, name string) lookup.StackScoped {
	stackScope := lookup.ParseStackScoped(workspace, name)
	stackScope.Version = ""
	return lookup.ParseStackScoped(workspace, stackScope.String())
}

func (r *Stage) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("must specify the service to update")
	}

	workspace, err := ctx.Workspace()
	if err != nil {
		return err
	}

	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	w, err := waiter.NewWaiter(ctx)
	if err != nil {
		return err
	}

	stackScope := stripRevision(workspace, ctx.CLI.Args()[0])

	resource, err := lookup.Lookup(ctx, stackScope.ResourceID, client.ServiceType)
	if err != nil {
		byID, err2 := wc.Service.ByID(ctx.CLI.Args()[0])
		if err2 != nil {
			return err
		}

		stackScope = stripRevision(workspace, lookup.StackScopedFromLabels(workspace, byID.Labels).String())
		resource = &types.NamedResource{
			Resource: byID.Resource,
			Name:     byID.Name,
		}
	}

	baseService, err := wc.Service.ByID(resource.ID)
	if err != nil {
		return err
	}

	revision, err := determineRevision(workspace, ctx.CLI.Args()[0], &resource.Resource)
	if err != nil {
		return err
	}

	stackScope.Version = revision
	_, err = lookup.Lookup(ctx, stackScope.String(), client.ServiceType)
	if err == nil {
		return fmt.Errorf("revision %s already exists", ctx.CLI.Args()[0])
	}

	args := append([]string{r.Image}, ctx.CLI.Args()[1:]...)
	serviceDef, err := r.ToService(args)
	if err != nil {
		return err
	}

	serviceDef.Name = fmt.Sprintf("%s-%s", resource.Name, revision)
	serviceDef.ParentService = resource.Name
	serviceDef.Version = revision
	serviceDef.Weight = int64(r.Weight)
	serviceDef.Scale = int64(r.Scale)
	serviceDef.SpaceID = baseService.SpaceID
	serviceDef.StackID = baseService.StackID
	if serviceDef.Scale == 0 {
		serviceDef.Scale = baseService.Scale
	}

	revService, err := wc.Service.Create(serviceDef)
	if err != nil {
		return err
	}

	w.Add(&revService.Resource)
	return w.Wait(ctx.Ctx)
}
