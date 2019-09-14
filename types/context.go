package types

import (
	"context"

	webhookinator "github.com/rancher/gitwatcher/pkg/generated/controllers/gitwatcher.cattle.io"
	"github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io"
	"github.com/rancher/rio/pkg/generated/controllers/autoscale.rio.cattle.io"
	"github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/apiextensions.k8s.io"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/apps"
	serving "github.com/rancher/wrangler-api/pkg/generated/controllers/autoscaling.internal.knative.dev"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/certmanager.k8s.io"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/core"
	networkingv1beta1 "github.com/rancher/wrangler-api/pkg/generated/controllers/networking"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/networking.istio.io"
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

type Context struct {
	Namespace string

	Apps          *apps.Factory
	AutoScale     *autoscale.Factory
	Build         *build.Factory
	CertManager   *certmanager.Factory
	Core          *core.Factory
	Ext           *apiextensions.Factory
	K8sNetworking *networkingv1beta1.Factory
	Networking    *networking.Factory
	Global        *admin.Factory
	K8s           kubernetes.Interface
	RBAC          *rbac.Factory
	Rio           *rio.Factory
	SMI           *smi.Factory
	Serving       *serving.Factory
	Storage       *storage.Factory
	Webhook       *webhookinator.Factory

	Apply apply.Apply
}

func From(ctx context.Context) *Context {
	return ctx.Value(contextKey{}).(*Context)
}

func NewContext(namespace string, config *rest.Config) *Context {
	context := &Context{
		Namespace:     namespace,
		Apps:          apps.NewFactoryFromConfigOrDie(config),
		AutoScale:     autoscale.NewFactoryFromConfigOrDie(config),
		Build:         build.NewFactoryFromConfigOrDie(config),
		CertManager:   certmanager.NewFactoryFromConfigOrDie(config),
		Core:          core.NewFactoryFromConfigOrDie(config),
		Ext:           apiextensions.NewFactoryFromConfigOrDie(config),
		K8sNetworking: networkingv1beta1.NewFactoryFromConfigOrDie(config),
		Networking:    networking.NewFactoryFromConfigOrDie(config),
		Global:        admin.NewFactoryFromConfigOrDie(config),
		RBAC:          rbac.NewFactoryFromConfigOrDie(config),
		Rio:           rio.NewFactoryFromConfigOrDie(config),
		Serving:       serving.NewFactoryFromConfigOrDie(config),
		Storage:       storage.NewFactoryFromConfigOrDie(config),
		SMI:           smi.NewFactoryFromConfigOrDie(config),
		Webhook:       webhookinator.NewFactoryFromConfigOrDie(config),
		K8s:           kubernetes.NewForConfigOrDie(config),
	}

	context.Apply = apply.New(context.K8s.Discovery(), apply.NewClientFactory(config))
	return context
}

func (c *Context) Start(ctx context.Context) error {
	return start.All(ctx, 5, c.Global)
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
