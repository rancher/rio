package v1beta1

import (
	"context"
	"sync"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"github.com/rancher/norman/objectclient/dynamic"
	"github.com/rancher/norman/restwatch"
	"k8s.io/client-go/rest"
)

type contextKeyType struct{}

type Interface interface {
	RESTClient() rest.Interface
	controller.Starter

	ListenConfigsGetter
	PublicDomainsGetter
}

type Client struct {
	sync.Mutex
	restClient rest.Interface
	starters   []controller.Starter

	listenConfigControllers map[string]ListenConfigController
	publicDomainControllers map[string]PublicDomainController
}

func Factory(ctx context.Context, config rest.Config) (context.Context, controller.Starter, error) {
	c, err := NewForConfig(config)
	if err != nil {
		return ctx, nil, err
	}

	return context.WithValue(ctx, contextKeyType{}, c), c, nil
}

func From(ctx context.Context) Interface {
	return ctx.Value(contextKeyType{}).(Interface)
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

		listenConfigControllers: map[string]ListenConfigController{},
		publicDomainControllers: map[string]PublicDomainController{},
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

type ListenConfigsGetter interface {
	ListenConfigs(namespace string) ListenConfigInterface
}

func (c *Client) ListenConfigs(namespace string) ListenConfigInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &ListenConfigResource, ListenConfigGroupVersionKind, listenConfigFactory{})
	return &listenConfigClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type PublicDomainsGetter interface {
	PublicDomains(namespace string) PublicDomainInterface
}

func (c *Client) PublicDomains(namespace string) PublicDomainInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &PublicDomainResource, PublicDomainGroupVersionKind, publicDomainFactory{})
	return &publicDomainClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}
