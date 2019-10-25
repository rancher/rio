package publicdomain

import (
	"fmt"

	"github.com/rancher/rio/pkg/services"

	"github.com/rancher/rio/cli/pkg/clicontext"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type Register struct {
	Version string `desc:"target to specific version instead of whole app"`
	Secret  string `desc:"use specified secret that contains TLS certs and key instead of build-in letsencrypt. The secret has to be created first in your system namespace(default: rio-system)"`
}

func (r *Register) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 2 {
		return fmt.Errorf("incorrect Usage. Example: `rio domain register DOMAIN_NAME TARGET`")
	}
	domainName := ctx.CLI.Args().Get(0)
	target := ctx.CLI.Args().Get(1)

	var targetApp, targetVersion, targetNamespace string
	result, err := ctx.ByID(target)
	if err != nil {
		return err
	}
	switch result.Object.(type) {
	case *riov1.Service:
		svc := result.Object.(*riov1.Service)
		targetApp, _ = services.AppAndVersion(svc)
		targetNamespace = svc.Namespace
	case *riov1.Router:
		router := result.Object.(*riov1.Router)
		targetApp, targetNamespace = router.Name, router.Namespace
	}

	targetVersion = r.Version

	pd := adminv1.PublicDomain{
		Spec: adminv1.PublicDomainSpec{
			SecretName:      r.Secret,
			TargetApp:       targetApp,
			TargetVersion:   targetVersion,
			TargetNamespace: targetNamespace,
		},
	}

	return ctx.Create(adminv1.NewPublicDomain(targetNamespace, domainName, pd))
}
