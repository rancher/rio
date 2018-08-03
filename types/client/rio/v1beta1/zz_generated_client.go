package client

import (
	"github.com/rancher/norman/clientbase"
)

type Client struct {
	clientbase.APIBaseClient

	Stack    StackOperations
	Service  ServiceOperations
	Config   ConfigOperations
	Volume   VolumeOperations
	RouteSet RouteSetOperations
}

func NewClient(opts *clientbase.ClientOpts) (*Client, error) {
	baseClient, err := clientbase.NewAPIClient(opts)
	if err != nil {
		return nil, err
	}

	client := &Client{
		APIBaseClient: baseClient,
	}

	client.Stack = newStackClient(client)
	client.Service = newServiceClient(client)
	client.Config = newConfigClient(client)
	client.Volume = newVolumeClient(client)
	client.RouteSet = newRouteSetClient(client)

	return client, nil
}
