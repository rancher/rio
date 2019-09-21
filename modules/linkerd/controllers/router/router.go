package router

import (
	"context"
	"fmt"
	"strconv"

	splitv1alpha1 "github.com/deislabs/smi-sdk-go/pkg/apis/split/v1alpha1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
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
	c := stackobject.NewGeneratingController(ctx, rContext, "routing-routers", rContext.Rio.Rio().V1().Router())
	c.Apply = c.Apply.WithStrictCaching().
		WithCacheTypes(rContext.SMI.Split().V1alpha1().TrafficSplit()).WithRateLimiting(10)

	sh := &routeHandler{
		systemNamespace:    rContext.Namespace,
		clusterDomainCache: rContext.Global.Admin().V1().ClusterDomain().Cache(),
		serviceCache:       rContext.Rio.Rio().V1().Service().Cache(),
		secrets:            rContext.Core.Core().V1().Secret(),
	}

	c.Populator = sh.populate
	return nil
}

type routeHandler struct {
	systemNamespace    string
	clusterDomainCache projectv1controller.ClusterDomainCache
	serviceCache       v1.ServiceCache
	secrets            corev1controller.SecretController
}

func (r routeHandler) populate(obj runtime.Object, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	router := obj.(*riov1.Router)

	if router == nil || router.DeletionTimestamp != nil {
		return nil
	}

	for i, route := range router.Spec.Routes {
		name := fmt.Sprintf("%s-%v", router.Name, i)
		split := createSplit(name, router, route)
		os.Add(split)
	}

	return nil
}

func createSplit(name string, router *riov1.Router, routerSpec riov1.RouteSpec) *splitv1alpha1.TrafficSplit {
	split := constructors.NewTrafficSplit(router.Namespace, name, splitv1alpha1.TrafficSplit{
		Spec: splitv1alpha1.TrafficSplitSpec{
			Service: name,
		},
	})
	for _, to := range routerSpec.To {
		if len(routerSpec.To) == 1 {
			to.Weight = 100
		}
		dest := to.Service
		if to.Revision != "" {
			dest = dest + "-" + to.Revision
		}
		split.Spec.Backends = append(split.Spec.Backends, splitv1alpha1.TrafficSplitBackend{
			Service: dest,
			Weight:  resource.MustParse(strconv.Itoa(to.Weight)),
		})
	}
	return split
}
