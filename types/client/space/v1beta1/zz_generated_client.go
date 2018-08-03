package client

import (
	"github.com/rancher/norman/clientbase"
)

type Client struct {
	clientbase.APIBaseClient

	ListenConfig ListenConfigOperations
	Node         NodeOperations
	Pod          PodOperations
	Space        SpaceOperations
}

func NewClient(opts *clientbase.ClientOpts) (*Client, error) {
	baseClient, err := clientbase.NewAPIClient(opts)
	if err != nil {
		return nil, err
	}

	client := &Client{
		APIBaseClient: baseClient,
	}

	client.ListenConfig = newListenConfigClient(client)
	client.Node = newNodeClient(client)
	client.Pod = newPodClient(client)
	client.Space = newSpaceClient(client)

	return client, nil
}
