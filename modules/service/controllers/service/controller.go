package service

import (
	"context"

	"github.com/rancher/rio/modules/service/controllers/service/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/arch"
	"github.com/rancher/rio/pkg/config"
	adminv1 "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
	corev1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/objectset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {

	sh := &serviceHandler{
		namespace:          rContext.Namespace,
		publicDomainCache:  rContext.Admin.Admin().V1().PublicDomain().Cache(),
		clusterDomainCache: rContext.Admin.Admin().V1().ClusterDomain().Cache(),
		configmaps:         rContext.Core.Core().V1().ConfigMap(),
	}

	riov1controller.RegisterServiceGeneratingHandler(ctx,
		rContext.Rio.Rio().V1().Service(),
		rContext.Apply.WithCacheTypes(
			rContext.RBAC.Rbac().V1().Role(),
			rContext.RBAC.Rbac().V1().RoleBinding(),
			rContext.Apps.Apps().V1().Deployment(),
			rContext.Apps.Apps().V1().DaemonSet(),
			rContext.Core.Core().V1().ServiceAccount(),
			rContext.Core.Core().V1().Service(),
			rContext.Core.Core().V1().Secret(),
			rContext.Core.Core().V1().PersistentVolumeClaim()).
			WithInjectorName("mesh").
			WithRateLimiting(20),
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
	configmaps         corev1controller.ConfigMapClient
}

func (s *serviceHandler) populate(service *riov1.Service, status riov1.ServiceStatus) ([]runtime.Object, riov1.ServiceStatus, error) {
	if service.Spec.Template {
		return nil, status, generic.ErrSkip
	}

	if err := s.ensureFeatures(service); err != nil {
		return nil, status, err
	}

	os := objectset.NewObjectSet()
	if err := populate.Service(service, os); err != nil {
		return nil, status, err
	}

	return os.All(), status, nil
}

func (s *serviceHandler) ensureFeatures(service *riov1.Service) error {
	cm, err := s.configmaps.Get(s.namespace, config.ConfigName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	conf, err := config.FromConfigMap(cm)
	if err != nil {
		return err
	}

	t := true
	if services.AutoscaleEnable(service) && arch.IsAmd64() {
		if conf.Features == nil {
			conf.Features = map[string]config.FeatureConfig{}
		}
		f := conf.Features["autoscaling"]
		f.Enabled = &t
		conf.Features["autoscaling"] = f
	}

	for _, con := range services.ToNamedContainers(service) {
		if con.ImageBuild != nil && con.ImageBuild.Repo != "" && arch.IsAmd64() {
			if conf.Features == nil {
				conf.Features = map[string]config.FeatureConfig{}
			}
			f := conf.Features["build"]
			f.Enabled = &t
			conf.Features["build"] = f
			break
		}
	}

	cm, err = config.SetConfig(cm, conf)
	if err != nil {
		return err
	}

	if _, err := s.configmaps.Update(cm); err != nil {
		return err
	}

	return nil
}
