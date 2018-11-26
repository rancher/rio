package istio

import (
	"context"

	"github.com/rancher/rio/types/apis/space.cattle.io/v1beta1"

	"github.com/rancher/norman/pkg/changeset"
	"github.com/rancher/rio/pkg/certs"
	"github.com/rancher/rio/pkg/deploy/istio"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	v12 "github.com/rancher/types/apis/core/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	all = "_istio_deploy_"
)

var trigger = []changeset.Key{
	{
		Name: all,
	},
}

func Register(ctx context.Context, rContext *types.Context) {
	s := &istioDeployController{
		virtualServiceLister: rContext.Networking.VirtualService.Cache(),
		serviceLister:        rContext.Core.Service.Cache(),
		namespaceLister:      rContext.Core.Namespace.Cache(),
		publicdomainLister:   rContext.Global.PublicDomain.Cache(),
		secrets:              rContext.Core.Secret.Cache(),
	}

	rContext.Networking.VirtualService.Interface().AddHandler(ctx, "istio-deploy", s.sync)
	changeset.Watch(ctx, "istio-deploy",
		resolve,
		rContext.Networking.VirtualService,
		rContext.Networking.VirtualService,
		rContext.Core.Service,
		rContext.Core.Namespace)
	rContext.Networking.VirtualService.Enqueue("", all)
}

func resolve(namespace, name string, obj runtime.Object) ([]changeset.Key, error) {
	switch t := obj.(type) {
	case *v1alpha3.VirtualService:
		return trigger, nil
	case *v1.Namespace:
		if t.Name == settings.IstioExternalLBNamespace {
			return trigger, nil
		}
	}

	return nil, nil
}

type istioDeployController struct {
	virtualServiceLister v1alpha3.VirtualServiceClientCache
	serviceLister        v12.ServiceClientCache
	namespaceLister      v12.NamespaceClientCache
	publicdomainLister   v1beta1.PublicDomainClientCache
	secrets              v12.SecretClientCache
}

func (i *istioDeployController) sync(key string, obj *v1alpha3.VirtualService) (runtime.Object, error) {
	if key != all {
		return nil, nil
	}

	lbNamespace, err := i.namespaceLister.Get("", settings.IstioExternalLBNamespace)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	vss, err := i.virtualServiceLister.List("", labels.Everything())
	if err != nil {
		return nil, err
	}

	pds, err := i.publicdomainLister.List("", labels.Everything())
	if err != nil {
		return nil, err
	}

	secret, err := i.secrets.Get(settings.IstioExternalLBNamespace, certs.TlsSecretName)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	return nil, istio.Deploy(lbNamespace, vss, pds, secret)
}
