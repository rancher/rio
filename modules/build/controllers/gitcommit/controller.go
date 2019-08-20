package gitcommit

import (
	"context"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	webhookv1controller "github.com/rancher/gitwatcher/pkg/generated/controllers/gitwatcher.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := Handler{
		ctx:              ctx,
		appsCache:        rContext.Rio.Rio().V1().App().Cache(),
		services:         rContext.Rio.Rio().V1().Service(),
		stacks:           rContext.Rio.Rio().V1().Stack(),
		gitWatcherCache:  rContext.Webhook.Gitwatcher().V1().GitWatcher().Cache(),
		gitWatcherClient: rContext.Webhook.Gitwatcher().V1().GitWatcher(),
	}

	wupdator := webhookv1controller.UpdateGitCommitOnChange(rContext.Webhook.Gitwatcher().V1().GitCommit().Updater(), h.onChange)
	rContext.Webhook.Gitwatcher().V1().GitCommit().OnChange(ctx, "webhook-execution", wupdator)

	return nil
}

type Handler struct {
	ctx              context.Context
	appsCache        riov1controller.AppCache
	gitWatcherCache  webhookv1controller.GitWatcherCache
	gitWatcherClient webhookv1controller.GitWatcherClient
	services         riov1controller.ServiceController
	stacks           riov1controller.StackController
}

func (h Handler) onChange(key string, obj *webhookv1.GitCommit) (*webhookv1.GitCommit, error) {
	if obj == nil {
		return obj, nil
	}

	gitWatcher, err := h.gitWatcherCache.Get(obj.Namespace, obj.Spec.GitWatcherName)
	if err != nil {
		return nil, err
	}

	if isOwnedByStack(gitWatcher) {
		return h.onChangeStack(key, obj, gitWatcher)
	}

	return h.onChangeService(key, obj, gitWatcher)
}

func isOwnedByStack(gitWatcher *webhookv1.GitWatcher) bool {
	return gitWatcher.Annotations["objectset.rio.cattle.io/owner-gvk"] == "rio.cattle.io/v1, Kind=Stack"
}
