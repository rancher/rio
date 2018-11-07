package types

import (
	"context"

	"github.com/rancher/norman"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	spacev1beta1 "github.com/rancher/rio/types/apis/space.cattle.io/v1beta1"
	appsv1 "github.com/rancher/types/apis/apps/v1beta2"
	"github.com/rancher/types/apis/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type contextKey struct{}

type Context struct {
	Apps        appsv1.Interface
	Core        v1.Interface
	Global      spacev1beta1.Interface
	K8s         kubernetes.Interface
	LocalConfig *rest.Config
	Networking  v1alpha3.Interface
	Rio         v1beta1.Interface
}

func Store(ctx context.Context, c *Context) context.Context {
	return context.WithValue(ctx, contextKey{}, c)
}

func From(ctx context.Context) *Context {
	return ctx.Value(contextKey{}).(*Context)
}

func NewContext(ctx context.Context) *Context {
	server := norman.GetServer(ctx)
	return &Context{
		Apps:        appsv1.From(ctx),
		Core:        v1.From(ctx),
		Global:      spacev1beta1.From(ctx),
		K8s:         server.K8sClient,
		LocalConfig: server.LocalConfig,
		Networking:  v1alpha3.From(ctx),
		Rio:         v1beta1.From(ctx),
	}
}

func BuildContext(ctx context.Context) (context.Context, error) {
	return Store(ctx, NewContext(ctx)), nil
}

func Register(f func(context.Context, *Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return f(ctx, From(ctx))
	}
}
