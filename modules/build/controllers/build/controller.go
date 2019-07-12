package build

import (
	"context"
	"errors"

	"github.com/knative/pkg/apis"
	"github.com/rancher/rio/modules/build/controllers/service"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	projectv1controller "github.com/rancher/rio/pkg/generated/controllers/admin.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	tektonv1alpha1controller "github.com/rancher/wrangler-api/pkg/generated/controllers/tekton.dev/v1alpha1"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		systemNamespace: rContext.Namespace,
		services:        rContext.Rio.Rio().V1().Service(),
	}

	rContext.Build.Tekton().V1alpha1().TaskRun().OnChange(ctx, "build-service-update", tektonv1alpha1controller.UpdateTaskRunOnChange(rContext.Build.Tekton().V1alpha1().TaskRun().Updater(), h.updateService))
	rContext.Build.Tekton().V1alpha1().TaskRun().OnRemove(ctx, "build-service-remove", h.updateServiceOnRemove)
	return nil
}

type handler struct {
	registry           string
	systemNamespace    string
	services           riov1controller.ServiceController
	clusterDomainCache projectv1controller.ClusterDomainCache
}

func (h handler) updateService(key string, build *tektonv1alpha1.TaskRun) (*tektonv1alpha1.TaskRun, error) {
	if build == nil {
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

	if build.IsSuccessful() {
		rev := svc.Spec.Build.Revision
		if rev == "" {
			rev = svc.Status.FirstRevision
		}
		imageName := service.PullImageName(rev, svc)
		if svc.Spec.Image != imageName {
			deepCopy := svc.DeepCopy()
			v1.ServiceConditionImageReady.SetError(deepCopy, "", nil)
			deepCopy.Spec.Image = service.PullImageName(rev, deepCopy)
			if _, err := h.services.Update(deepCopy); err != nil {
				return build, err
			}
		}
	} else if build.IsDone() {
		con := build.Status.GetCondition(apis.ConditionSucceeded)
		deepCopy := svc.DeepCopy()
		v1.ServiceConditionImageReady.SetError(deepCopy, con.Reason, errors.New(con.Message))
		_, err := h.services.Update(deepCopy)
		return build, err
	}

	return build, nil
}

func (h *handler) updateServiceOnRemove(key string, build *tektonv1alpha1.TaskRun) (*tektonv1alpha1.TaskRun, error) {
	if build == nil {
		return build, nil
	}

	if !build.IsDone() {
		namespace := build.Labels["service-namespace"]
		name := build.Labels["service-name"]
		h.services.Enqueue(namespace, name)
	}

	return build, nil
}
