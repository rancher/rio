package service

import (
	"context"
	"github.com/rancher/wrangler/pkg/generic"

	"github.com/rancher/rio/modules/service/controllers/service/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	adminv1 "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/objectset"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	sc := rContext.Rio.Rio().V1().Service()
	sh := &serviceHandler{
		namespace:          rContext.Namespace,
		publicDomainCache:  rContext.Admin.Admin().V1().PublicDomain().Cache(),
		clusterDomainCache: rContext.Admin.Admin().V1().ClusterDomain().Cache(),
	}

	riov1controller.RegisterServiceGeneratingHandler(ctx,
		sc,
		rContext.Apply.WithCacheTypes(
			rContext.RBAC.Rbac().V1().Role(),
			rContext.RBAC.Rbac().V1().RoleBinding(),
			rContext.Apps.Apps().V1().Deployment(),
			rContext.Apps.Apps().V1().DaemonSet(),
			rContext.Core.Core().V1().ServiceAccount(),
			rContext.Core.Core().V1().Service(),
			rContext.Core.Core().V1().Secret()).
			WithRateLimiting(5).
			WithInjectorName("mesh").
			WithStrictCaching(),
		"ServiceDeployed",
		"service",
		sh.populate,
		nil)

	return nil
}

type serviceHandler struct {
	namespace string

	clusterDomainCache adminv1.ClusterDomainCache
	publicDomainCache  adminv1.PublicDomainCache
}

func (s *serviceHandler) populate(service *riov1.Service, status riov1.ServiceStatus) ([]runtime.Object, riov1.ServiceStatus, error) {
	if service.Spec.Template {
		return nil, status, generic.ErrSkip
	}

	os := objectset.NewObjectSet()
	if err := populate.Service(service, os); err != nil {
		return nil, status, err
	}

	return os.All(), status, nil
}
