package istio

import (
	"context"

	"github.com/rancher/norman/pkg/changeset"
	"github.com/rancher/rio/pkg/certs"
	"github.com/rancher/rio/pkg/deploy/istio"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	"github.com/rancher/rio/types/apis/space.cattle.io/v1beta1"
	v12 "github.com/rancher/types/apis/core/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		virtualServiceLister: rContext.Networking.VirtualServices("").Controller().Lister(),
		serviceLister:        rContext.Core.Services("").Controller().Lister(),
		namespaceLister:      rContext.Core.Namespaces("").Controller().Lister(),
		publicdomainLister:   rContext.Global.PublicDomains("").Controller().Lister(),
		secrets:              rContext.Core.Secrets(settings.IstioExternalLBNamespace),
	}

	rContext.Networking.VirtualServices("").AddHandler(ctx, "istio-deploy", s.sync)
	changeset.Watch(ctx, "istio-deploy",
		resolve,
		rContext.Networking.VirtualServices("").Controller().Enqueue,
		rContext.Networking.VirtualServices("").Controller(),
		rContext.Core.Services("").Controller(),
		rContext.Core.Namespaces("").Controller())
	rContext.Networking.VirtualServices("").Controller().Enqueue("", all)
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
	virtualServiceLister v1alpha3.VirtualServiceLister
	serviceLister        v12.ServiceLister
	namespaceLister      v12.NamespaceLister
	publicdomainLister   v1beta1.PublicDomainLister
	secrets              v12.SecretInterface
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

	secret, err := i.secrets.Get(certs.TlsSecretName, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	return nil, istio.Deploy(lbNamespace, vss, pds, secret)
}
