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

	ServiceScaleRecommendationsGetter
}

type Clients struct {
	Interface Interface

	ServiceScaleRecommendation ServiceScaleRecommendationClient
}

type Client struct {
	sync.Mutex
	restClient rest.Interface
	starters   []controller.Starter

	serviceScaleRecommendationControllers map[string]ServiceScaleRecommendationController
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

		ServiceScaleRecommendation: &serviceScaleRecommendationClient2{
			iface: iface.ServiceScaleRecommendations(""),
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

		serviceScaleRecommendationControllers: map[string]ServiceScaleRecommendationController{},
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

type ServiceScaleRecommendationsGetter interface {
	ServiceScaleRecommendations(namespace string) ServiceScaleRecommendationInterface
}

func (c *Client) ServiceScaleRecommendations(namespace string) ServiceScaleRecommendationInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &ServiceScaleRecommendationResource, ServiceScaleRecommendationGroupVersionKind, serviceScaleRecommendationFactory{})
	return &serviceScaleRecommendationClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}
