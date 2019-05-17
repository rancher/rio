package publicdomain

import (
	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
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

	namespace, name := stack.NamespaceAndName(ctx, target)

	return ctx.Create(adminv1.NewPublicDomain(namespace, name2.PublicDomain(domainName), adminv1.PublicDomain{
		Spec: adminv1.PublicDomainSpec{
			DomainName:        domainName,
			TargetServiceName: name,
		},
	}))
}
