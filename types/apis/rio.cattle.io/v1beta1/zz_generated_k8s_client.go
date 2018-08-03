package v1beta1

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

	StacksGetter
	ServicesGetter
	ConfigsGetter
	VolumesGetter
	RouteSetsGetter
}

type Client struct {
	sync.Mutex
	restClient rest.Interface
	starters   []controller.Starter

	stackControllers    map[string]StackController
	serviceControllers  map[string]ServiceController
	configControllers   map[string]ConfigController
	volumeControllers   map[string]VolumeController
	routeSetControllers map[string]RouteSetController
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

		stackControllers:    map[string]StackController{},
		serviceControllers:  map[string]ServiceController{},
		configControllers:   map[string]ConfigController{},
		volumeControllers:   map[string]VolumeController{},
		routeSetControllers: map[string]RouteSetController{},
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

type StacksGetter interface {
	Stacks(namespace string) StackInterface
}

func (c *Client) Stacks(namespace string) StackInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &StackResource, StackGroupVersionKind, stackFactory{})
	return &stackClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type ServicesGetter interface {
	Services(namespace string) ServiceInterface
}

func (c *Client) Services(namespace string) ServiceInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &ServiceResource, ServiceGroupVersionKind, serviceFactory{})
	return &serviceClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type ConfigsGetter interface {
	Configs(namespace string) ConfigInterface
}

func (c *Client) Configs(namespace string) ConfigInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &ConfigResource, ConfigGroupVersionKind, configFactory{})
	return &configClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type VolumesGetter interface {
	Volumes(namespace string) VolumeInterface
}

func (c *Client) Volumes(namespace string) VolumeInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &VolumeResource, VolumeGroupVersionKind, volumeFactory{})
	return &volumeClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type RouteSetsGetter interface {
	RouteSets(namespace string) RouteSetInterface
}

func (c *Client) RouteSets(namespace string) RouteSetInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &RouteSetResource, RouteSetGroupVersionKind, routeSetFactory{})
	return &routeSetClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}
