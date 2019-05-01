package types

import (
	"context"

	apiextensions "github.com/rancher/rio/pkg/generated/controllers/apiextensions.k8s.io"
	"github.com/rancher/rio/pkg/generated/controllers/apps"
	autoscale "github.com/rancher/rio/pkg/generated/controllers/autoscale.rio.cattle.io"
	build "github.com/rancher/rio/pkg/generated/controllers/build.knative.dev"
	certmanager "github.com/rancher/rio/pkg/generated/controllers/certmanager.k8s.io"
	"github.com/rancher/rio/pkg/generated/controllers/core"
	"github.com/rancher/rio/pkg/generated/controllers/extensions"
	git "github.com/rancher/rio/pkg/generated/controllers/git.rio.cattle.io"
	networking "github.com/rancher/rio/pkg/generated/controllers/networking.istio.io"
	project "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io"
	"github.com/rancher/rio/pkg/generated/controllers/rbac"
	rio "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io"
	"github.com/rancher/rio/pkg/generated/controllers/storage"
	webhookinator "github.com/rancher/rio/pkg/generated/controllers/webhookinator.rio.cattle.io"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/start"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type contextKey struct{}

type Context struct {
	Namespace      string
	CustomRegistry string

	Apps        *apps.Factory
	AutoScale   *autoscale.Factory
	Build       *build.Factory
	CertManager *certmanager.Factory
	Core        *core.Factory
	Ext         *apiextensions.Factory
	Extensions  *extensions.Factory
	Global      *project.Factory
	Git         *git.Factory
	K8s         kubernetes.Interface
	Networking  *networking.Factory
	RBAC        *rbac.Factory
	Rio         *rio.Factory
	Storage     *storage.Factory
	Webhook     *webhookinator.Factory

	Apply apply.Apply
}

func From(ctx context.Context) *Context {
	return ctx.Value(contextKey{}).(*Context)
}

func NewContext(namespace string, config *rest.Config) *Context {
	context := &Context{
		Namespace:   namespace,
		Apps:        apps.NewFactoryFromConfigOrDie(config),
		AutoScale:   autoscale.NewFactoryFromConfigOrDie(config),
		Build:       build.NewFactoryFromConfigOrDie(config),
		CertManager: certmanager.NewFactoryFromConfigOrDie(config),
		Core:        core.NewFactoryFromConfigOrDie(config),
		Ext:         apiextensions.NewFactoryFromConfigOrDie(config),
		Extensions:  extensions.NewFactoryFromConfigOrDie(config),
		Global:      project.NewFactoryFromConfigOrDie(config),
		Git:         git.NewFactoryFromConfigOrDie(config),
		Networking:  networking.NewFactoryFromConfigOrDie(config),
		RBAC:        rbac.NewFactoryFromConfigOrDie(config),
		Rio:         rio.NewFactoryFromConfigOrDie(config),
		Storage:     storage.NewFactoryFromConfigOrDie(config),
		Webhook:     webhookinator.NewFactoryFromConfigOrDie(config),
		K8s:         kubernetes.NewForConfigOrDie(config),
	}

	context.Apply = apply.New(context.K8s.Discovery(), apply.NewClientFactory(config))
	return context
}

func (c *Context) Start(ctx context.Context) error {
	return start.All(ctx, 5,
		c.Apps,
		c.AutoScale,
		c.Build,
		c.CertManager,
		c.Core,
		c.Extensions,
		c.Ext,
		c.Global,
		c.Git,
		c.Networking,
		c.RBAC,
		c.Rio,
		c.Storage,
		c.Webhook,
	)
}

func BuildContext(ctx context.Context, namespace string, registry string, config *rest.Config) (context.Context, *Context) {
	c := NewContext(namespace, config)
	c.CustomRegistry = registry
	return context.WithValue(ctx, contextKey{}, c), c
}

func Register(f func(context.Context, *Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return f(ctx, From(ctx))
	}
}
