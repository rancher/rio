package ps

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

type ServiceData struct {
	ID       string
	Created  string
	Service  *riov1.Service
	Stack    *riov1.Stack
	Endpoint string
	External string
}

func (p *Ps) services(ctx *clicontext.CLIContext, stacks map[string]bool) error {
	services, err := ctx.List(types.ServiceType)
	if err != nil {
		return err
	}

	externalServices, err := ctx.List(types.ExternalServiceType)
	if err != nil {
		return err
	}

	// routes
	routes, err := ctx.List(types.ExternalServiceType)
	if err != nil {
		return err
	}

	all := append(services, externalServices...)
	all = append(all, routes...)

	writer := tables.NewService(ctx)
	return writer.Write(all)
}
