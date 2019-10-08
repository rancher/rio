package gitcommit

import (
	"context"
	"fmt"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	webhookv1controller "github.com/rancher/gitwatcher/pkg/generated/controllers/gitwatcher.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := Handler{
		ctx:              ctx,
		namespace:        rContext.Namespace,
		services:         rContext.Rio.Rio().V1().Service(),
		stacks:           rContext.Rio.Rio().V1().Stack(),
		gitWatcherCache:  rContext.Webhook.Gitwatcher().V1().GitWatcher().Cache(),
		gitWatcherClient: rContext.Webhook.Gitwatcher().V1().GitWatcher(),
		gitcommits:       rContext.Webhook.Gitwatcher().V1().GitCommit(),
	}

	rContext.Webhook.Gitwatcher().V1().GitCommit().OnChange(ctx, "webhook-execution", h.onChange)

	rContext.Rio.Rio().V1().Service().OnChange(ctx, "service-update-gitcommit", h.updateGitcommit)

	return nil
}

type Handler struct {
	ctx              context.Context
	namespace        string
	gitWatcherCache  webhookv1controller.GitWatcherCache
	gitWatcherClient webhookv1controller.GitWatcherClient
	gitcommits       webhookv1controller.GitCommitController
	services         riov1controller.ServiceController
	stacks           riov1controller.StackController
}

func (h Handler) onChange(key string, obj *webhookv1.GitCommit) (*webhookv1.GitCommit, error) {
	if obj == nil {
		return obj, nil
	}

	if webhookv1.GitWebHookExecutionConditionHandled.IsTrue(obj) {
		return obj, nil
	}

	gitWatcher, err := h.gitWatcherCache.Get(obj.Namespace, obj.Spec.GitWatcherName)
	if err != nil {
		return nil, err
	}

	if isOwnedByStack(gitWatcher) {
		if _, err := h.onChangeStack(key, obj, gitWatcher); err != nil {
			return nil, err
		}
	}

	if _, err := h.onChangeService(key, obj, gitWatcher); err != nil {
		return nil, err
	}

	obj = obj.DeepCopy()
	webhookv1.GitWebHookExecutionConditionHandled.SetStatus(obj, "True")
	_, err = h.gitcommits.Update(obj)
	return obj, err
}

func (h Handler) updateGitcommit(key string, obj *riov1.Service) (*riov1.Service, error) {
	if obj == nil || obj.DeletionTimestamp != nil {
		return obj, nil
	}

	if obj.Status.GitCommitName == "" {
		return obj, nil
	}

	rev := obj.Spec.ImageBuild.Revision
	if rev == "" {
		return obj, nil
	}

	gitcommit, err := h.gitcommits.Cache().Get(obj.Namespace, obj.Status.GitCommitName)
	if err != nil {
		return obj, err
	}

	if gitcommit.Status.GithubStatus == nil {
		return obj, nil
	}

	webhook, err := h.services.Cache().Get(h.namespace, "webhook")
	if err != nil {
		return obj, err
	}
	gitcommit = gitcommit.DeepCopy()
	logURL := fmt.Sprintf("%s/logs/%s/%s?log-token=%s", webhook.Status.Endpoints[0], obj.Namespace, obj.Name, obj.Status.BuildLogToken)
	endpoint := ""
	if len(obj.Status.Endpoints) > 0 {
		endpoint = obj.Status.Endpoints[0]
	}
	state := "in_progress"
	if obj.Status.DeploymentReady {
		state = "success"
	}
	update := false
	if gitcommit.Status.GithubStatus.LogURL != logURL || gitcommit.Status.GithubStatus.EnvironmentURL != endpoint || gitcommit.Status.GithubStatus.DeploymentState != state {
		update = true
	}
	if !update {
		return obj, nil
	}

	gitcommit.Status.GithubStatus.LogURL = logURL
	gitcommit.Status.GithubStatus.EnvironmentURL = endpoint
	gitcommit.Status.GithubStatus.DeploymentState = state
	if _, err := h.gitcommits.Update(gitcommit); err != nil {
		return obj, err
	}
	return obj, nil
}

func isOwnedByStack(gitWatcher *webhookv1.GitWatcher) bool {
	return gitWatcher.Annotations["objectset.rio.cattle.io/owner-gvk"] == "rio.cattle.io/v1, Kind=Stack"
}
