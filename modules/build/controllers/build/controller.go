package build

import (
	"context"
	"errors"

	webhookv1controller "github.com/rancher/gitwatcher/pkg/generated/controllers/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/modules/build/controllers/service"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"knative.dev/pkg/apis"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		systemNamespace: rContext.Namespace,
		services:        rContext.Rio.Rio().V1().Service(),
		stacks:          rContext.Rio.Rio().V1().Stack(),
		gitcommits:      rContext.Webhook.Gitwatcher().V1().GitCommit(),
	}

	rContext.Build.Tekton().V1alpha1().TaskRun().OnChange(ctx, "build-service-update", h.updateService)
	rContext.Build.Tekton().V1alpha1().TaskRun().OnRemove(ctx, "build-service-remove", h.updateOnRemove)
	return nil
}

type handler struct {
	systemNamespace string
	services        riov1controller.ServiceController
	stacks          riov1controller.StackController
	gitcommits      webhookv1controller.GitCommitController
}

func (h handler) updateService(key string, build *tektonv1alpha1.TaskRun) (*tektonv1alpha1.TaskRun, error) {
	if build == nil {
		return build, nil
	}

	namespace := build.Labels["service-namespace"]
	name := build.Labels["service-name"]
	containerName := build.Labels["container-name"]
	svc, err := h.services.Cache().Get(namespace, name)
	if err != nil {
		return build, nil
	}

	if svc.Spec.Image != "" {
		return build, nil
	}

	state := ""
	if build.IsDone() && !build.IsSuccessful() {
		state = "failure"
	} else if !build.IsDone() {
		state = "in_progress"
	}
	if build.Labels["gitcommit-name"] != "" {
		gitcommit, err := h.gitcommits.Cache().Get(build.Namespace, build.Labels["gitcommit-name"])
		if err != nil {
			return build, err
		}
		gitcommit = gitcommit.DeepCopy()
		if gitcommit.Status.BuildStatus != state {
			gitcommit.Status.BuildStatus = state
			if _, err := h.gitcommits.Update(gitcommit); err != nil {
				return build, err
			}
		}
	}

	if build.IsSuccessful() {
		containers := []v1.NamedContainer{
			{
				Name:      name,
				Container: svc.Spec.Container,
			},
		}
		containers = append(containers, svc.Spec.Sidecars...)

		for _, con := range containers {
			if con.Name == containerName {
				rev := con.ImageBuild.Revision
				imageName := service.PullImageName(rev, namespace, containerName, con.ImageBuild)
				if svc.Spec.Image != imageName {
					deepCopy := svc.DeepCopy()
					v1.ServiceConditionImageReady.SetError(deepCopy, "", nil)
					deepCopy.Spec.Image = service.PullImageName(rev, namespace, containerName, con.ImageBuild)
					if _, err := h.services.Update(deepCopy); err != nil {
						return build, err
					}
					if _, err := h.services.UpdateStatus(deepCopy); err != nil {
						return build, err
					}
				}
			}
		}
	} else if build.IsDone() {
		con := build.Status.GetCondition(apis.ConditionSucceeded)
		deepCopy := svc.DeepCopy()
		v1.ServiceConditionImageReady.SetError(deepCopy, con.Reason, errors.New(con.Message))
		if _, err := h.services.Update(deepCopy); err != nil {
			return build, err
		}
		_, err := h.services.UpdateStatus(deepCopy)
		return build, err
	}

	return build, nil
}

func (h *handler) updateOnRemove(key string, build *tektonv1alpha1.TaskRun) (*tektonv1alpha1.TaskRun, error) {
	if build == nil {
		return build, nil
	}

	if !build.IsSuccessful() || !build.IsDone() {
		if build.Labels["service-namespace"] != "" {
			namespace := build.Labels["service-namespace"]
			name := build.Labels["service-name"]
			h.services.Enqueue(namespace, name)
		} else if build.Labels["stack-namespace"] != "" {
			namespace := build.Labels["stack-namespace"]
			name := build.Labels["stack-name"]
			h.stacks.Enqueue(namespace, name)
		}
	}

	return build, nil
}
