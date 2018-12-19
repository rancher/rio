package v1

import (
	"context"
	"sync"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"github.com/rancher/norman/objectclient/dynamic"
	"github.com/rancher/norman/restwatch"
	"k8s.io/client-go/rest"
)

type (
	contextKeyType        struct{}
	contextClientsKeyType struct{}
)

type Interface interface {
	RESTClient() rest.Interface
	controller.Starter

	StacksGetter
	ExternalServicesGetter
	ServicesGetter
	ConfigsGetter
	VolumesGetter
	RouteSetsGetter
}

type Clients struct {
	Interface Interface

	Stack           StackClient
	ExternalService ExternalServiceClient
	Service         ServiceClient
	Config          ConfigClient
	Volume          VolumeClient
	RouteSet        RouteSetClient
}

type Client struct {
	sync.Mutex
	restClient rest.Interface
	starters   []controller.Starter

	stackControllers           map[string]StackController
	externalServiceControllers map[string]ExternalServiceController
	serviceControllers         map[string]ServiceController
	configControllers          map[string]ConfigController
	volumeControllers          map[string]VolumeController
	routeSetControllers        map[string]RouteSetController
}

func Factory(ctx context.Context, config rest.Config) (context.Context, controller.Starter, error) {
	c, err := NewForConfig(config)
	if err != nil {
		return ctx, nil, err
	}

	cs := NewClientsFromInterface(c)

	ctx = context.WithValue(ctx, contextKeyType{}, c)
	ctx = context.WithValue(ctx, contextClientsKeyType{}, cs)
	return ctx, c, nil
}

func ClientsFrom(ctx context.Context) *Clients {
	return ctx.Value(contextClientsKeyType{}).(*Clients)
}

func From(ctx context.Context) Interface {
	return ctx.Value(contextKeyType{}).(Interface)
}

func NewClients(config rest.Config) (*Clients, error) {
	iface, err := NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return NewClientsFromInterface(iface), nil
}

func NewClientsFromInterface(iface Interface) *Clients {
	return &Clients{
		Interface: iface,

		Stack: &stackClient2{
			iface: iface.Stacks(""),
		},
		ExternalService: &externalServiceClient2{
			iface: iface.ExternalServices(""),
		},
		Service: &serviceClient2{
			iface: iface.Services(""),
		},
		Config: &configClient2{
			iface: iface.Configs(""),
		},
		Volume: &volumeClient2{
			iface: iface.Volumes(""),
		},
		RouteSet: &routeSetClient2{
			iface: iface.RouteSets(""),
		},
	}
}

func NewForConfig(config rest.Config) (Interface, error) {
	if config.NegotiatedSerializer == nil {
		config.NegotiatedSerializer = dynamic.NegotiatedSerializer
	}

	restClient, err := restwatch.UnversionedRESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &Client{
		restClient: restClient,

		stackControllers:           map[string]StackController{},
		externalServiceControllers: map[string]ExternalServiceController{},
		serviceControllers:         map[string]ServiceController{},
		configControllers:          map[string]ConfigController{},
		volumeControllers:          map[string]VolumeController{},
		routeSetControllers:        map[string]RouteSetController{},
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

type ExternalServicesGetter interface {
	ExternalServices(namespace string) ExternalServiceInterface
}

func (c *Client) ExternalServices(namespace string) ExternalServiceInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &ExternalServiceResource, ExternalServiceGroupVersionKind, externalServiceFactory{})
	return &externalServiceClient{
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
