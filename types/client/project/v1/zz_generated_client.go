package client

import (
	"github.com/rancher/norman/clientbase"
)

type Client struct {
	clientbase.APIBaseClient

	ListenConfig ListenConfigOperations
	Node         NodeOperations
	Pod          PodOperations
	Project      ProjectOperations
	PublicDomain PublicDomainOperations
	Feature      FeatureOperations
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
	client.Project = newProjectClient(client)
	client.PublicDomain = newPublicDomainClient(client)
	client.Feature = newFeatureClient(client)

	return client, nil
}
