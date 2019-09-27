package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/rio/modules/gateway/controllers/service/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-serviceset", rContext.Rio.Rio().V1().App())
	c.Apply = c.Apply.WithStrictCaching().
		WithCacheTypes(rContext.Networking.Networking().V1alpha3().DestinationRule(),
			rContext.K8sNetworking.Extensions().V1beta1().Ingress(),
			rContext.Networking.Networking().V1alpha3().VirtualService()).WithRateLimiting(10)

	sh := &serviceHandler{
		systemNamespace:    rContext.Namespace,
		serviceCache:       rContext.Rio.Rio().V1().Service().Cache(),
		secretCache:        rContext.Core.Core().V1().Secret().Cache(),
		clusterDomainCache: rContext.Global.Admin().V1().ClusterDomain().Cache(),
	}

	c.Populator = sh.populate
	return nil
}

type serviceHandler struct {
	systemNamespace    string
	serviceCache       v1.ServiceCache
	secretCache        corev1controller.SecretCache
	clusterDomainCache projectv1controller.ClusterDomainCache
}

func (s serviceHandler) populate(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	app := obj.(*riov1.App)
	if app == nil {
		return nil
	}

	clusterDomain, err := s.clusterDomainCache.Get(s.systemNamespace, constants.ClusterDomainName)
	if err != nil {
		return err
	}

	if len(app.Spec.Revisions) == 0 {
		return nil
	}

	dr := destinationRuleForService(app)
	os.Add(dr)

	public := false
	for _, rev := range app.Spec.Revisions {
		if rev.Public {
			public = true
		}
	}
	if !public {
		return nil
	}

	var dests []populate.Dest
	for version, rev := range app.Status.RevisionWeight {
		dests = append(dests, populate.Dest{
			Host:   app.Name,
			Subset: version,
			Weight: rev.Weight,
		})
	}
	sort.Slice(dests, func(i, j int) bool {
		return dests[i].Subset < dests[j].Subset
	})

	var revision *riov1.Service
	for i := len(app.Spec.Revisions) - 1; i >= 0; i-- {
		revision, err = s.serviceCache.Get(app.Namespace, app.Spec.Revisions[i].ServiceName)
		if err != nil && !errors.IsNotFound(err) {
			return err
		} else if errors.IsNotFound(err) {
			continue
		}
		break
	}
	if revision == nil {
		return nil
	}

	deepcopy := revision.DeepCopy()
	deepcopy.Status.PublicDomains = app.Status.PublicDomains
	revVs := populate.VirtualServiceFromSpec(true, s.systemNamespace, app.Name, app.Namespace, clusterDomain, deepcopy, dests...)
	os.Add(revVs)

	return nil
}

func destinationRuleForService(app *riov1.App) *v1alpha3.DestinationRule {
	drSpec := v1alpha3.DestinationRuleSpec{
		Host: fmt.Sprintf("%s.%s.svc.cluster.local", app.Name, app.Namespace),
	}

	for _, rev := range app.Spec.Revisions {
		drSpec.Subsets = append(drSpec.Subsets, newSubSet(rev.Version))
	}

	dr := newDestinationRule(app.Namespace, app.Name)
	dr.Spec = drSpec

	return dr
}

func newSubSet(version string) v1alpha3.Subset {
	return v1alpha3.Subset{
		Name: version,
		Labels: map[string]string{
			"version": version,
		},
	}
}

func newDestinationRule(namespace, name string) *v1alpha3.DestinationRule {
	return constructors.NewDestinationRule(namespace, name, v1alpha3.DestinationRule{})
}
