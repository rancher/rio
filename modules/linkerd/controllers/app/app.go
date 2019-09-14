package app

import (
	"context"
	"fmt"
	"strconv"

	splitv1alpha1 "github.com/deislabs/smi-sdk-go/pkg/apis/split/v1alpha1"
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
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-serviceset", rContext.Rio.Rio().V1().App())
	c.Apply = c.Apply.WithStrictCaching().
		WithCacheTypes(rContext.K8sNetworking.Networking().V1beta1().Ingress(),
			rContext.SMI.Split().V1alpha1().TrafficSplit()).WithRateLimiting(10)

	sh := &appHandler{
		systemNamespace:    rContext.Namespace,
		clusterDomainCache: rContext.Global.Admin().V1().ClusterDomain().Cache(),
		serviceCache:       rContext.Rio.Rio().V1().Service().Cache(),
		secrets:            rContext.Core.Core().V1().Secret(),
	}

	c.Populator = sh.populate
	return nil
}

type appHandler struct {
	systemNamespace    string
	clusterDomainCache projectv1controller.ClusterDomainCache
	serviceCache       v1.ServiceCache
	secrets            corev1controller.SecretController
}

func (a appHandler) populate(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	app := obj.(*riov1.App)

	clusterDomain, err := a.clusterDomainCache.Get(a.systemNamespace, constants.ClusterDomainName)
	if err != nil {
		return err
	}

	if clusterDomain.Status.ClusterDomain == "" {
		return nil
	}

	if app.Namespace != a.systemNamespace {
		split := constructors.NewTrafficSplit(app.Namespace, app.Name, splitv1alpha1.TrafficSplit{
			Spec: splitv1alpha1.TrafficSplitSpec{
				Service: app.Name,
			},
		})
		for ver, rev := range app.Status.RevisionWeight {
			split.Spec.Backends = append(split.Spec.Backends, splitv1alpha1.TrafficSplitBackend{
				Service: fmt.Sprintf("%s-%s", app.Name, ver),
				Weight:  resource.MustParse(strconv.Itoa(rev.Weight)),
			})
		}
		os.Add(split)
	}
	return nil
}
