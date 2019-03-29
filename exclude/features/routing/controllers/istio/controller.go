package istio

import (
	"context"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/features/letsencrypt/controllers/issuer"
	"github.com/rancher/rio/features/routing/controllers/istio/populate"
	corev1controller "github.com/rancher/rio/pkg/generated/controllers/core/v1"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/objectset"
	"github.com/rancher/wrangler/pkg/relatedresource"
	"github.com/rancher/wrangler/pkg/trigger"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	evalTrigger trigger.Trigger
)

func ReevalIstio() {
	if evalTrigger != nil {
		evalTrigger.Trigger()
	}
}

func Register(ctx context.Context, rContext *types.Context) error {
	s := &istioDeployController{
		systemNamespace: rContext.SystemNamespace,
		gatewayApply: rContext.Apply.WithSetID("istio-stack").
			WithCacheTypes(rContext.Networking.Networking().V1alpha3().Gateway()),
		stackApply: rContext.Apply.WithSetID("istio-gateway").
			WithCacheTypes(rContext.Rio.Rio().V1().Stack()),
		publicDomainLister: rContext.Global.Project().V1().PublicDomain().Cache(),
		secretsLister:      rContext.Core.Core().V1().Secret().Cache(),
	}

	evalTrigger = trigger.New(rContext.Networking.Networking().V1alpha3().VirtualService())
	evalTrigger.OnTrigger(ctx, "istio-deploy", s.sync)

	relatedresource.Watch(ctx, "istio-deploy",
		resolve,
		rContext.Networking.Networking().V1alpha3().VirtualService(),
		rContext.Networking.Networking().V1alpha3().VirtualService(),
		rContext.Core.Core().V1().Namespace())

	return nil
}

func resolve(namespace, name string, obj runtime.Object) ([]relatedresource.Key, error) {
	switch t := obj.(type) {
	case *v1alpha3.VirtualService:
		return []relatedresource.Key{evalTrigger.Key()}, nil
	case *v1.Namespace:
		if t.Name == settings.IstioStackName {
			return []relatedresource.Key{evalTrigger.Key()}, nil
		}
	}

	return nil, nil
}

type istioDeployController struct {
	systemNamespace    string
	gatewayApply       apply.Apply
	stackApply         apply.Apply
	publicDomainLister projectv1controller.PublicDomainCache
	secretsLister      corev1controller.SecretCache
}

func (i *istioDeployController) sync() error {
	output := objectset.NewObjectSet()
	if err := populate.PopulateStack(i.systemNamespace, output); err != nil {
		output.AddErr(err)
	}
	if err := i.stackApply.Apply(output); err != nil {
		return err
	}

	pds, err := i.publicDomainLister.List("", labels.Everything())
	if err != nil {
		return err
	}

	secret, err := i.secretsLister.Get(settings.IstioStackName, issuer.TLSSecretName)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	os := populate.Istio(i.systemNamespace, pds, secret)
	return i.gatewayApply.Apply(os)
}
