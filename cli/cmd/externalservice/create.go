package externalservice

import (
	"fmt"
	"net"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type Create struct {
}

func (c *Create) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) < 2 {
		return errors.New("Incorrect usage. Example: `rio externalservice create NAME TARGET...`")
	}

	var externalService riov1.ExternalService

	for i, name := range ctx.CLI.Args().Tail() {
		switch ip := net.ParseIP(name); {
		case ip != nil:
			externalService.Spec.IPAddresses = append(externalService.Spec.IPAddresses, name)
		case strings.ContainsRune(name, '.'):
			externalService.Spec.FQDN = name
		default:
			externalService.Spec.Service = name
		}

		if i > 0 && len(externalService.Spec.IPAddresses) != (i+1) {
			return fmt.Errorf("multiple targets is for IP addresses only")
		}
	}

	namespace, name := stack.NamespaceAndName(ctx, ctx.CLI.Args()[0])

	return ctx.Create(riov1.NewExternalService(namespace, name, externalService))
}
