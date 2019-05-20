package build

import (
	"context"
	"errors"

	"github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/rancher/rio/modules/build/controllers/service"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	v1alpha12 "github.com/rancher/wrangler-api/pkg/generated/controllers/build.knative.dev/v1alpha1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		systemNamespace:    rContext.Namespace,
		services:           rContext.Rio.Rio().V1().Service(),
		clusterDomainCache: rContext.Global.Admin().V1().ClusterDomain().Cache(),
	}

	rContext.Build.Build().V1alpha1().Build().OnChange(ctx, "build-service-update", v1alpha12.UpdateBuildOnChange(rContext.Build.Build().V1alpha1().Build().Updater(), h.updateService))
	rContext.Build.Build().V1alpha1().Build().OnRemove(ctx, "build-service-remove", h.updateServiceOnRemove)
	return nil
}

type handler struct {
	registry           string
	systemNamespace    string
	services           riov1controller.ServiceController
	clusterDomainCache projectv1controller.ClusterDomainCache
}

func (h handler) updateService(key string, build *v1alpha1.Build) (*v1alpha1.Build, error) {
	if build == nil {
		return build, nil
	}

	clusterDomain, err := h.clusterDomainCache.Get(h.systemNamespace, constants.ClusterDomainName)
	if err != nil {
		return build, err
	}
	domain := clusterDomain.Status.ClusterDomain
	if domain == "" {
		return build, nil
	}

	con := build.Status.GetCondition(v1alpha1.BuildSucceeded)
	if con == nil {
		return build, nil
	}

	namespace := build.Labels["service-namespace"]
	name := build.Labels["service-name"]
	svc, err := h.services.Cache().Get(namespace, name)
	if err != nil {
		return build, nil
	}

	if svc.Spec.Image != "" {
		return build, nil
	}

	if con.IsTrue() {
		imageName := service.ImageName(h.registry, h.systemNamespace, build.Spec.Source.Git.Revision, domain, svc)
		if svc.Spec.Image != imageName {
			deepCopy := svc.DeepCopy()
			v1.ServiceConditionImageReady.SetError(deepCopy, "", nil)
			deepCopy.Spec.Image = service.ImageName(h.registry, h.systemNamespace, build.Spec.Source.Git.Revision, domain, deepCopy)
			if _, err := h.services.Update(deepCopy); err != nil {
				return build, err
			}
		}
	} else if con.IsFalse() {
		deepCopy := svc.DeepCopy()
		v1.ServiceConditionImageReady.SetError(deepCopy, con.Reason, errors.New(con.Message))
		_, err := h.services.Update(deepCopy)
		return build, err
	}

	return build, nil
}

func (h *handler) updateServiceOnRemove(key string, build *v1alpha1.Build) (*v1alpha1.Build, error) {
	if build == nil {
		return build, nil
	}

	con := build.Status.GetCondition(v1alpha1.BuildSucceeded)
	if con != nil && con.IsFalse() {
		namespace := build.Labels["service-namespace"]
		name := build.Labels["service-name"]
		h.services.Enqueue(namespace, name)
	}

	return build, nil
}
