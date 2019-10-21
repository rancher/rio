package publicdomain

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/clicontext"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	"k8s.io/apimachinery/pkg/api/errors"
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
		if !errors.IsNotFound(err) {
			return err
		}
		targetApp = target
		targetNamespace = ctx.GetSetNamespace()
	} else {
		svc := result.Object.(*riov1.Service)
		targetApp, targetVersion = services.AppAndVersion(svc)
		targetNamespace = svc.Namespace
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
