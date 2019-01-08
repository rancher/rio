package types

import (
	"context"

	"github.com/rancher/norman"
	"github.com/rancher/norman/controller"
	"github.com/rancher/rio/types/apis/apiextensions.k8s.io/v1beta1"
	cmv1alpha1 "github.com/rancher/rio/types/apis/certmanager.k8s.io/v1alpha1"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	policyv1beta1 "github.com/rancher/rio/types/apis/policy/v1beta1"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	autoscalev1 "github.com/rancher/rio/types/apis/rio-autoscale.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	storagev1 "github.com/rancher/rio/types/apis/storage.k8s.io/v1"
	appsv1 "github.com/rancher/types/apis/apps/v1beta2"
	"github.com/rancher/types/apis/core/v1"
	rbacv1 "github.com/rancher/types/apis/rbac.authorization.k8s.io/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type contextKey struct{}

type Context struct {
	InCluster   bool
	Apps        *appsv1.Clients
	AutoScale   *autoscalev1.Clients
	CertManager *cmv1alpha1.Clients
	Core        *v1.Clients
	Ext         *v1beta1.Clients
	Global      *projectv1.Clients
	K8s         kubernetes.Interface
	LocalConfig *rest.Config
	Networking  *v1alpha3.Clients
	Policy      *policyv1beta1.Clients
	RBAC        *rbacv1.Clients
	Rio         *riov1.Clients
	Storage     *storagev1.Clients
}

func (c *Context) Starters() []controller.Starter {
	return []controller.Starter{
		c.Apps.Interface,
		c.CertManager.Interface,
		c.Core.Interface,
		c.Ext.Interface,
		c.Global.Interface,
		c.Networking.Interface,
		c.Policy.Interface,
		c.RBAC.Interface,
		c.Rio.Interface,
		c.Storage.Interface,
	}
}

func From(ctx context.Context) *Context {
	return ctx.Value(contextKey{}).(*Context)
}

func NewContext(ctx context.Context) *Context {
	server := norman.GetServer(ctx)
	return &Context{
		Apps:        appsv1.ClientsFrom(ctx),
		AutoScale:   autoscalev1.ClientsFrom(ctx),
		CertManager: cmv1alpha1.ClientsFrom(ctx),
		Core:        v1.ClientsFrom(ctx),
		Ext:         v1beta1.ClientsFrom(ctx),
		Global:      projectv1.ClientsFrom(ctx),
		K8s:         server.K8sClient,
		LocalConfig: server.LocalConfig,
		Networking:  v1alpha3.ClientsFrom(ctx),
		Policy:      policyv1beta1.ClientsFrom(ctx),
		RBAC:        rbacv1.ClientsFrom(ctx),
		Rio:         riov1.ClientsFrom(ctx),
		Storage:     storagev1.ClientsFrom(ctx),
	}
}

func BuildContext(ctx context.Context) (context.Context, error) {
	c := NewContext(ctx)
	return context.WithValue(ctx, contextKey{}, c), nil
}

func Register(f func(context.Context, *Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return f(ctx, From(ctx))
	}
}
