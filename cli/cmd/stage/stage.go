package stage

import (
	"fmt"

	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
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
	return nil
}
