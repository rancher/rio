package build

import (
	"context"
	"errors"

	webhookv1controller "github.com/rancher/gitwatcher/pkg/generated/controllers/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/modules/build/controllers/service"
	"github.com/rancher/rio/modules/build/pkg"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/condition"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		systemNamespace: rContext.Namespace,
		services:        rContext.Rio.Rio().V1().Service(),
		stacks:          rContext.Rio.Rio().V1().Stack(),
		gitcommits:      rContext.Webhook.Gitwatcher().V1().GitCommit(),
	}

	rContext.Build.Tekton().V1alpha1().TaskRun().OnChange(ctx, "build-service-update", h.updateService)
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

	namespace, svcName, conName := build.Namespace, build.Labels[pkg.ServiceLabel], build.Labels[pkg.ContainerLabel]
	svc, err := h.services.Cache().Get(namespace, svcName)
	if err != nil {
		return build, nil
	}

	if svc.Spec.Template {
		return build, nil
	}

	state := ""
	if condition.Cond("Succeeded").IsFalse(build) {
		state = "failure"
	} else if condition.Cond("Succeeded").IsUnknown(build) {
		state = "in_progress"
	}

	if build.Labels[pkg.GitCommitLabel] != "" {
		gitcommit, err := h.gitcommits.Cache().Get(build.Namespace, build.Labels[pkg.GitCommitLabel])
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

	if condition.Cond("Succeeded").IsTrue(build) {
		if conName == services.RootContainerName(svc) {
			rev := svc.Spec.ImageBuild.Revision
			imageName := service.PullImageName(rev, namespace, conName, svc.Spec.ImageBuild)
			if svc.Spec.Image != imageName {
				deepCopy := svc.DeepCopy()
				v1.ServiceConditionImageReady.SetError(deepCopy, "", nil)
				deepCopy.Spec.Image = service.PullImageName(rev, namespace, conName, svc.Spec.ImageBuild)
				if _, err := h.services.Update(deepCopy); err != nil {
					return build, err
				}
			}
		} else {
			for i, con := range svc.Spec.Sidecars {
				if con.Name == conName {
					rev := con.ImageBuild.Revision
					imageName := service.PullImageName(rev, namespace, conName, con.ImageBuild)
					if con.Image != imageName {
						deepCopy := svc.DeepCopy()
						v1.ServiceConditionImageReady.SetError(deepCopy, "", nil)
						deepCopy.Spec.Sidecars[i].Image = service.PullImageName(rev, namespace, conName, con.ImageBuild)
						if _, err := h.services.Update(deepCopy); err != nil {
							return build, err
						}
					}
				}
			}
		}
	} else if condition.Cond("Succeeded").IsFalse(build) {
		reason := condition.Cond("Succeeded").GetReason(build)
		message := condition.Cond("Succeeded").GetMessage(build)

		deepCopy := svc.DeepCopy()
		v1.ServiceConditionImageReady.SetError(deepCopy, reason, errors.New(message))
		_, err := h.services.UpdateStatus(deepCopy)
		return build, err
	}

	return build, nil
}
