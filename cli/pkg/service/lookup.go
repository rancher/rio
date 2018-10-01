package service

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func Lookup(ctx *clicontext.CLIContext, name string) (*client.Service, error) {
	resource, err := lookup.Lookup(ctx.ClientLookup, name, client.ServiceType)
	if err != nil {
		return nil, err
	}

	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return nil, err
	}

	service, err := wc.Service.ByID(resource.ID)
	if err != nil {
		return nil, err
	}

	parsedService := lookup.ParseServiceName(name)
	if parsedService.Revision == "" || service.Version == parsedService.Revision {
		return service, nil
	}

	return wc.Service.ByID(resource.ID + "-" + parsedService.Revision)
}
