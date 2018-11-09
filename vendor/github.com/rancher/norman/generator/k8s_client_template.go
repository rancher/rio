package generator

var k8sClientTemplate = `package {{.version.Version}}

import (
	"sync"
	"context"

	"github.com/rancher/norman/objectclient"
	"github.com/rancher/norman/objectclient/dynamic"
	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/restwatch"
	"k8s.io/client-go/rest"
)

type contextKeyType struct{}

type Interface interface {
	RESTClient() rest.Interface
	controller.Starter
	{{range .schemas}}
	{{.CodeNamePlural}}Getter{{end}}
}

type Client struct {
	sync.Mutex
	restClient         rest.Interface
	starters           []controller.Starter
	{{range .schemas}}
	{{.ID}}Controllers map[string]{{.CodeName}}Controller{{end}}
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
		restClient:         restClient,
	{{range .schemas}}
		{{.ID}}Controllers: map[string]{{.CodeName}}Controller{},{{end}}
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

{{range .schemas}}
type {{.CodeNamePlural}}Getter interface {
	{{.CodeNamePlural}}(namespace string) {{.CodeName}}Interface
}

func (c *Client) {{.CodeNamePlural}}(namespace string) {{.CodeName}}Interface {
	objectClient := objectclient.NewObjectClient(namespace, c.restClient, &{{.CodeName}}Resource, {{.CodeName}}GroupVersionKind, {{.ID}}Factory{})
	return &{{.ID}}Client{
		ns:           namespace,
		client:       c,
		objectClient: objectClient,
	}
}
{{end}}
`
