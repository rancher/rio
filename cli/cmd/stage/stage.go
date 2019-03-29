package stage

import (
	"fmt"

	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/constants"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type Stage struct {
	create.Create
	Scale  int    `desc:"Number of replicas to run"`
	Weight int    `desc:"Percentage of traffic routed to staged service"`
	Image  string `desc:"Runtime image (Docker image/OCI image)"`
}

func determineRevision(ctx *clicontext.CLIContext, name string, service types.Resource) (string, error) {
	revision := "next"
	if name == service.Name {
		return revision, nil
	}

	parsedService := lookup.ParseStackScoped(ctx.GetDefaultStackName(), name)
	if parsedService.Version == constants.DefaultServiceVersion {
		return "", fmt.Errorf("\"%s\" is not a valid revision to stage", constants.DefaultServiceVersion)
	}
	if parsedService.Version != "" {
		revision = parsedService.Version
	}

	return revision, nil
}

func stripRevision(ctx *clicontext.CLIContext, name string) lookup.StackScoped {
	stackScope := lookup.ParseStackScoped(ctx.GetDefaultStackName(), name)
	stackScope.Version = ""
	return stackScope
}

func (r *Stage) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("must specify the service to update")
	}

	stackScope := stripRevision(ctx, ctx.CLI.Args()[0])

	resource, err := lookup.Lookup(ctx, stackScope.String(), types.ServiceType)
	if err != nil {
		return err
	}
	baseService := resource.Object.(*v1.Service)

	revision, err := determineRevision(ctx, ctx.CLI.Args()[0], resource)
	if err != nil {
		return err
	}

	stackScope.Version = revision
	_, err = lookup.Lookup(ctx, stackScope.String(), types.ServiceType)
	if err == nil {
		return fmt.Errorf("revision %s already exists", ctx.CLI.Args()[0])
	}

	args := append([]string{r.Image}, ctx.CLI.Args()[1:]...)
	serviceDef, err := r.ToService(args)
	if err != nil {
		return err
	}

	serviceDef.Name = fmt.Sprintf("%s-%s", resource.Name, revision)
	serviceDef.Namespace = baseService.Namespace
	serviceDef.Spec.Revision.ParentService = resource.Name
	serviceDef.Spec.Revision.Version = revision
	serviceDef.Spec.Revision.Weight = r.Weight
	serviceDef.Spec.Scale = r.Scale
	serviceDef.Spec.PortBindings = baseService.Spec.PortBindings
	if serviceDef.Spec.Scale == 0 {
		serviceDef.Spec.Scale = baseService.Spec.Scale
	}

	return ctx.Create(serviceDef)
}
