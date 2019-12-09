package types

import (
	"context"

	webhookinator "github.com/rancher/gitwatcher/pkg/generated/controllers/gitwatcher.cattle.io"
	"github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io"
	gateway "github.com/rancher/rio/pkg/generated/controllers/gateway.solo.io"
	gloo "github.com/rancher/rio/pkg/generated/controllers/gloo.solo.io"
	"github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/apiextensions.k8s.io"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/apps"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/batch"
	certmanager "github.com/rancher/wrangler-api/pkg/generated/controllers/cert-manager.io"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/core"
	extensionsv1beta1 "github.com/rancher/wrangler-api/pkg/generated/controllers/extensions"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/rbac"
	smi "github.com/rancher/wrangler-api/pkg/generated/controllers/split.smi-spec.io"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/storage"
	build "github.com/rancher/wrangler-api/pkg/generated/controllers/tekton.dev"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/start"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type contextKey struct{}

type Config interface {
	Get(section, name string) string
}

type Context struct {
	Namespace string

	Apps          *apps.Factory
	Batch         *batch.Factory
	Build         *build.Factory
	CertManager   *certmanager.Factory
	Core          *core.Factory
	Ext           *apiextensions.Factory
	K8sNetworking *extensionsv1beta1.Factory
	Admin         *admin.Factory
	K8s           kubernetes.Interface
	RBAC          *rbac.Factory
	Rio           *rio.Factory
	SMI           *smi.Factory
	Gateway       *gateway.Factory
	Gloo          *gloo.Factory
	Storage       *storage.Factory
	Webhook       *webhookinator.Factory

	RestConfig *rest.Config
	Config     Config
	Apply      apply.Apply
}

func From(ctx context.Context) *Context {
	return ctx.Value(contextKey{}).(*Context)
}

func NewContext(namespace string, config *rest.Config) *Context {
	context := &Context{
		Namespace:     namespace,
		Apps:          apps.NewFactoryFromConfigOrDie(config),
		Batch:         batch.NewFactoryFromConfigOrDie(config),
		Build:         build.NewFactoryFromConfigOrDie(config),
		CertManager:   certmanager.NewFactoryFromConfigOrDie(config),
		Core:          core.NewFactoryFromConfigOrDie(config),
		Ext:           apiextensions.NewFactoryFromConfigOrDie(config),
		K8sNetworking: extensionsv1beta1.NewFactoryFromConfigOrDie(config),
		Admin:         admin.NewFactoryFromConfigOrDie(config),
		RBAC:          rbac.NewFactoryFromConfigOrDie(config),
		Rio:           rio.NewFactoryFromConfigOrDie(config),
		Storage:       storage.NewFactoryFromConfigOrDie(config),
		SMI:           smi.NewFactoryFromConfigOrDie(config),
		Webhook:       webhookinator.NewFactoryFromConfigOrDie(config),
		K8s:           kubernetes.NewForConfigOrDie(config),
		Gateway:       gateway.NewFactoryFromConfigOrDie(config),
		Gloo:          gloo.NewFactoryFromConfigOrDie(config),
		RestConfig:    config,
	}

	context.Apply = apply.New(context.K8s.Discovery(), apply.NewClientFactory(config)).WithRateLimiting(20.0)
	return context
}

func (c *Context) Start(ctx context.Context) error {
	return start.All(ctx, 5,
		c.Apps,
		c.Batch,
		c.Build,
		c.CertManager,
		c.Core,
		c.Ext,
		c.K8sNetworking,
		c.Admin,
		c.RBAC,
		c.Rio,
		c.Storage,
		c.SMI,
		c.Webhook,
		c.Gateway,
		c.Gloo)
}

func BuildContext(ctx context.Context, namespace string, config *rest.Config) (context.Context, *Context) {
	c := NewContext(namespace, config)
	return context.WithValue(ctx, contextKey{}, c), c
}

func Register(f func(context.Context, *Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return f(ctx, From(ctx))
	}
}
