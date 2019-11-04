package externalservice

import (
	"fmt"
	"net"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

type Create struct {
}

func (c *Create) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) < 2 {
		return errors.New("Incorrect usage. Example: `rio externalservice create NAME TARGET...`")
	}

	var externalService riov1.ExternalService

	for i, name := range ctx.CLI.Args().Tail() {
		host, _ := kv.Split(name, ":")
		switch ip := net.ParseIP(host); {
		case ip != nil:
			externalService.Spec.IPAddresses = append(externalService.Spec.IPAddresses, name)
		case strings.ContainsRune(name, '.'):
			externalService.Spec.FQDN = name
		default:
			ns, name := kv.Split(name, "/")
			if name == "" {
				name = ns
				ns = ctx.GetSetNamespace()
			}
			externalService.Spec.TargetServiceName = name
			externalService.Spec.TargetServiceNamespace = ns
		}

		if i > 0 && len(externalService.Spec.IPAddresses) != (i+1) {
			return fmt.Errorf("multiple targets is for IP addresses only")
		}
	}

	r := ctx.ParseID(ctx.CLI.Args()[0])
	return ctx.Create(riov1.NewExternalService(r.Namespace, r.Name, externalService))
}
