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
	Apps        *appsv1.Clients
	Core        *v1.Clients
	Global      *spacev1beta1.Clients
	K8s         kubernetes.Interface
	LocalConfig *rest.Config
	Networking  *v1alpha3.Clients
	Rio         *v1beta1.Clients
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
		Apps:        appsv1.ClientsFrom(ctx),
		Core:        v1.ClientsFrom(ctx),
		Global:      spacev1beta1.ClientsFrom(ctx),
		K8s:         server.K8sClient,
		LocalConfig: server.LocalConfig,
		Networking:  v1alpha3.ClientsFrom(ctx),
		Rio:         v1beta1.ClientsFrom(ctx),
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
