package publicdomain

import (
	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	name2 "github.com/rancher/rio/pkg/name"
	v1 "k8s.io/api/core/v1"
)

type Register struct {
	Secret string `desc:"use specified secret that contains TLS certs and key instead of build-in letsencrypt. The secret has to be created first in your system namespace(default: rio-system)"`
}

func (a *Register) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 2 {
		return errors.New("Incorrect Usage. Example: `rio domain register DOMAIN_NAME TARGET_SVC`")
	}
	domainName := ctx.CLI.Args().Get(0)
	target := ctx.CLI.Args().Get(1)

	namespace, name := stack.NamespaceAndName(ctx, target)
	pd := adminv1.PublicDomain{
		Spec: adminv1.PublicDomainSpec{
			DomainName:        domainName,
			TargetServiceName: name,
		},
	}
	if a.Secret != "" {
		pd.Spec.SecretRef = v1.SecretReference{
			Namespace: ctx.SystemNamespace,
			Name:      a.Secret,
		}
		pd.Spec.DisableLetsencrypt = true
	}

	return ctx.Create(adminv1.NewPublicDomain(namespace, name2.PublicDomain(domainName), pd))
}
