package v1alpha3

import (
	"context"
	"sync"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"github.com/rancher/norman/restwatch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Interface interface {
	RESTClient() rest.Interface
	controller.Starter

	GatewaysGetter
	VirtualServicesGetter
}

type Client struct {
	sync.Mutex
	restClient rest.Interface
	starters   []controller.Starter

	gatewayControllers        map[string]GatewayController
	virtualServiceControllers map[string]VirtualServiceController
}

func NewForConfig(config rest.Config) (Interface, error) {
	if config.NegotiatedSerializer == nil {
		configConfig := dynamic.ContentConfig()
		config.NegotiatedSerializer = configConfig.NegotiatedSerializer
	}

	restClient, err := restwatch.UnversionedRESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &Client{
		restClient: restClient,

		gatewayControllers:        map[string]GatewayController{},
		virtualServiceControllers: map[string]VirtualServiceController{},
	}, nil
}

func (c *Client) RESTClient() rest.Interface {
	return c.restClient
}

func (c *Client) Sync(ctx context.Context) error {
	return controller.Sync(ctx, c.starters...)
}

func (c *Client) Start(ctx context.Context, threadiness int) error {
	return controller.Start(ctx, threadiness, c.starters...)
}

type GatewaysGetter interface {
	Gateways(namespace string) GatewayInterface
}

func (c *Client) Gateways(namespace string) GatewayInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &GatewayResource, GatewayGroupVersionKind, gatewayFactory{})
	return &gatewayClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type VirtualServicesGetter interface {
	VirtualServices(namespace string) VirtualServiceInterface
}

func (c *Client) VirtualServices(namespace string) VirtualServiceInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &VirtualServiceResource, VirtualServiceGroupVersionKind, virtualServiceFactory{})
	return &virtualServiceClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}
