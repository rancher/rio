package externalservice

import (
	"context"

	"github.com/rancher/rio/exclude/pkg/settings"

	v12 "github.com/rancher/rio/pkg/generated/controllers/project.rio.cattle.io/v1"

	"github.com/rancher/rio/modules/istio/controllers/externalservice/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/serviceset"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-external-service", rContext.Rio.Rio().V1().ExternalService())
	c.Apply.WithCacheTypes(rContext.Networking.Networking().V1alpha3().ServiceEntry(),
		rContext.Networking.Networking().V1alpha3().VirtualService())

	p := populator{
		namespace:          rContext.Namespace,
		serviceCache:       rContext.Rio.Rio().V1().Service().Cache(),
		clusterDomainCache: rContext.Global.Project().V1().ClusterDomain().Cache(),
	}

	c.Populator = p.populate
	return nil
}

type populator struct {
	namespace          string
	serviceCache       v1.ServiceCache
	clusterDomainCache v12.ClusterDomainCache
}

func (p populator) populate(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	if err := populate.ServiceEntry(obj.(*riov1.ExternalService), os); err != nil {
		return err
	}

	if obj.(*riov1.ExternalService).Spec.Service == "" {
		return nil
	}

	targetStackName, targetServiceName := kv.Split(obj.(*riov1.ExternalService).Spec.Service, "/")
	svc, err := p.serviceCache.Get(targetStackName, targetServiceName)
	if err != nil {
		return err
	}

	serviceSets, err := serviceset.CollectionServices([]*riov1.Service{svc})
	if err != nil {
		return err
	}

	serviceSet, ok := serviceSets[svc.Name]
	if !ok {
		return err
	}

	clusterDomain, err := p.clusterDomainCache.Get(p.namespace, settings.ClusterDomainName)

	populate.VirtualServiceForExternalService(p.namespace, obj.(*riov1.ExternalService), serviceSet, clusterDomain, svc, os)
	return nil
}
