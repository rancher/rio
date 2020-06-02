package types

import (
	"context"

	webhookinator "github.com/rancher/gitwatcher/pkg/generated/controllers/gitwatcher.cattle.io"
	config2 "github.com/rancher/rio/pkg/config"
	"github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io"
	"github.com/rancher/rio/pkg/generated/controllers/gateway.solo.io"
	"github.com/rancher/rio/pkg/generated/controllers/gloo.solo.io"
	istio "github.com/rancher/rio/pkg/generated/controllers/networking.istio.io"
	"github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io"
	smi "github.com/rancher/rio/pkg/generated/controllers/split.smi-spec.io"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/apiextensions.k8s.io"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/apps"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/batch"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/core"
	extensionsv1beta1 "github.com/rancher/wrangler-api/pkg/generated/controllers/extensions"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/rbac"
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
	Core          *core.Factory
	Ext           *apiextensions.Factory
	Istio         *istio.Factory
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

		RestConfig: config,
	}
	if config2.ConfigController.MeshMode == "istio" {
		context.Istio = istio.NewFactoryFromConfigOrDie(config)
	} else if config2.ConfigController.MeshMode == "linkerd" {
		context.Gloo = gloo.NewFactoryFromConfigOrDie(config)
		context.Gateway = gateway.NewFactoryFromConfigOrDie(config)
	}

	context.Apply = apply.New(context.K8s.Discovery(), apply.NewClientFactory(config)).WithRateLimiting(20.0)
	return context
}

func (c *Context) Start(ctx context.Context) error {
	starters := []start.Starter{
		c.Apps,
		c.Batch,
		c.Build,
		c.Core,
		c.Ext,
		c.K8sNetworking,
		c.Admin,
		c.RBAC,
		c.Rio,
		c.Storage,
		c.SMI,
		c.Webhook,
	}
	if config2.ConfigController.MeshMode == "istio" {
		starters = append(starters, c.Istio)
	} else if config2.ConfigController.MeshMode == "linkerd" {
		starters = append(starters, c.Gloo)
		starters = append(starters, c.Gateway)
	}
	return start.All(ctx, 5, starters...)
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
