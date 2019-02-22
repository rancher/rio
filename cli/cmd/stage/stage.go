package stage

import (
	"fmt"

	"github.com/rancher/rio/cli/cmd/create"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/types"
	"github.com/rancher/rio/pkg/settings"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Stage struct {
	create.Create
	Scale  int    `desc:"Number of replicas to run"`
	Weight int    `desc:"Percentage of traffic routed to staged service"`
	Image  string `desc:"Runtime image (Docker image/OCI image)"`
}

func determineRevision(project *clientcfg.Project, name string, service types.Resource) (string, error) {
	revision := "next"
	if name == service.Name {
		return revision, nil
	}

	parsedService := lookup.ParseStackScoped(project, name)
	if parsedService.Version == settings.DefaultServiceVersion {
		return "", fmt.Errorf("\"%s\" is not a valid revision to stage", settings.DefaultServiceVersion)
	}
	if parsedService.Version != "" {
		revision = parsedService.Version
	}

	return revision, nil
}

func stripRevision(project *clientcfg.Project, name string) lookup.StackScoped {
	stackScope := lookup.ParseStackScoped(project, name)
	stackScope.Version = ""
	return lookup.ParseStackScoped(project, stackScope.String())
}

func (r *Stage) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("must specify the service to update")
	}

	client, err := ctx.KubeClient()
	if err != nil {
		return err
	}

	project, err := ctx.Project()
	if err != nil {
		return err
	}

	stackScope := stripRevision(project, ctx.CLI.Args()[0])

	resource, err := lookup.Lookup(ctx, stackScope.ResourceID, types.ServiceType)
	if err != nil {
		return err
	}

	baseService, err := client.Rio.Services(project.Project.Name).Get(resource.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	revision, err := determineRevision(project, ctx.CLI.Args()[0], resource)
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
	serviceDef.Spec.Revision.ParentService = resource.Name
	serviceDef.Spec.Revision.Version = revision
	serviceDef.Spec.Revision.Weight = r.Weight
	serviceDef.Spec.Scale = r.Scale
	serviceDef.Spec.ProjectName = baseService.Spec.ProjectName
	serviceDef.Spec.StackName = baseService.Spec.StackName
	serviceDef.Spec.PortBindings = baseService.Spec.PortBindings
	if serviceDef.Spec.Scale == 0 {
		serviceDef.Spec.Scale = baseService.Spec.Scale
	}

	if _, err := client.Rio.Services(project.Project.Name).Create(serviceDef); err != nil {
		return err
	}

	return nil
}
