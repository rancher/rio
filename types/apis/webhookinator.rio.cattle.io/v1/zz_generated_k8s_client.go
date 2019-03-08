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

	GitWebHookReceiversGetter
	GitWebHookExecutionsGetter
}

type Clients struct {
	Interface Interface

	GitWebHookReceiver  GitWebHookReceiverClient
	GitWebHookExecution GitWebHookExecutionClient
}

type Client struct {
	sync.Mutex
	restClient rest.Interface
	starters   []controller.Starter

	gitWebHookReceiverControllers  map[string]GitWebHookReceiverController
	gitWebHookExecutionControllers map[string]GitWebHookExecutionController
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

		GitWebHookReceiver: &gitWebHookReceiverClient2{
			iface: iface.GitWebHookReceivers(""),
		},
		GitWebHookExecution: &gitWebHookExecutionClient2{
			iface: iface.GitWebHookExecutions(""),
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

		gitWebHookReceiverControllers:  map[string]GitWebHookReceiverController{},
		gitWebHookExecutionControllers: map[string]GitWebHookExecutionController{},
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

type GitWebHookReceiversGetter interface {
	GitWebHookReceivers(namespace string) GitWebHookReceiverInterface
}

func (c *Client) GitWebHookReceivers(namespace string) GitWebHookReceiverInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &GitWebHookReceiverResource, GitWebHookReceiverGroupVersionKind, gitWebHookReceiverFactory{})
	return &gitWebHookReceiverClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}

type GitWebHookExecutionsGetter interface {
	GitWebHookExecutions(namespace string) GitWebHookExecutionInterface
}

func (c *Client) GitWebHookExecutions(namespace string) GitWebHookExecutionInterface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &GitWebHookExecutionResource, GitWebHookExecutionGroupVersionKind, gitWebHookExecutionFactory{})
	return &gitWebHookExecutionClient{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}
