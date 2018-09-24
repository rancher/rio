package service

import (
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func Lookup(ctx *server.Context, name string) (*client.Service, error) {
	resource, err := lookup.Lookup(ctx.ClientLookup, name, client.ServiceType)
	if err != nil {
		return nil, err
	}

	service, err := ctx.Client.Service.ByID(resource.ID)
	if err != nil {
		return nil, err
	}

	parsedService := lookup.ParseServiceName(name)
	if parsedService.Revision == "" || service.Version == parsedService.Revision {
		return service, nil
	}

	return ctx.Client.Service.ByID(resource.ID + "-" + parsedService.Revision)
}
