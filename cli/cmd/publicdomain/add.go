package publicdomain

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/kv"
)

type Register struct {
	Secret string `desc:"use specified secret that contains TLS certs and key instead of build-in letsencrypt. The secret has to be created first in your system namespace(default: rio-system)"`
}

func (r *Register) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 2 {
		return fmt.Errorf("incorrect Usage. Example: `rio domain register DOMAIN_NAME TARGET`")
	}
	domainName := ctx.CLI.Args().Get(0)
	target := ctx.CLI.Args().Get(1)

	pd := adminv1.PublicDomain{
		Spec: adminv1.PublicDomainSpec{
			SecretName: r.Secret,
		},
	}

	/*
		1) if target is [namespace:]name and can't be found as an object, treat it as app
		2) if target is [namespace:]name@version, treat it as service(revision)
		3) if target is [namespace:]name and can be found as an object, treat it as a router
	*/
	result, err := ctx.ByID(target)
	if err != nil {
		ns, app := kv.RSplit(target, ":")
		if ns == "" {
			ns = ctx.GetSetNamespace()
		}
		pd.Spec.TargetApp = app
		pd.Spec.TargetNamespace = ns
	} else {
		switch result.Object.(type) {
		case *riov1.Service:
			svc := result.Object.(*riov1.Service)
			targetApp, targetVersion := services.AppAndVersion(svc)
			if !strings.Contains(target, "@") {
				targetVersion = ""
			}
			pd.Spec.TargetApp = targetApp
			pd.Spec.TargetVersion = targetVersion
			pd.Spec.TargetNamespace = svc.Namespace
		case *riov1.Router:
			router := result.Object.(*riov1.Router)
			pd.Spec.TargetRouter = router.Name
			pd.Spec.TargetNamespace = router.Namespace
		}
	}

	return ctx.Create(adminv1.NewPublicDomain(pd.Spec.TargetNamespace, domainName, pd))
}
