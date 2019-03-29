package publicdomain

import (
	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	name2 "github.com/rancher/rio/pkg/name"
)

type Add struct {
}

func (a *Add) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 2 {
		return errors.New("Incorrect Usage. Example: `rio domain add DOMAIN_NAME TARGET_SVC`")
	}
	domainName := ctx.CLI.Args().Get(0)
	target := ctx.CLI.Args().Get(1)

	_, namespace, name, err := stack.ResolveSpaceStackForName(ctx, target)
	if err != nil {
		return err
	}

	return ctx.Create(projectv1.NewPublicDomain(ctx.Namespace, name2.PublicDomain(domainName), projectv1.PublicDomain{
		Spec: projectv1.PublicDomainSpec{
			DomainName:      domainName,
			TargetStackName: namespace,
			TargetName:      name,
		},
	}))
}
