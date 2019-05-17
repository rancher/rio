package gitcommit

import (
	"context"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	webhookv1controller "github.com/rancher/gitwatcher/pkg/generated/controllers/gitwatcher.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := Handler{
		ctx:             ctx,
		appsCache:       rContext.Rio.Rio().V1().App().Cache(),
		services:        rContext.Rio.Rio().V1().Service(),
		gitWatcherCache: rContext.Webhook.Gitwatcher().V1().GitWatcher().Cache(),
	}

	wupdator := webhookv1controller.UpdateGitCommitOnChange(rContext.Webhook.Gitwatcher().V1().GitCommit().Updater(), h.onChange)
	rContext.Webhook.Gitwatcher().V1().GitCommit().OnChange(ctx, "webhook-execution", wupdator)

	return nil
}

type Handler struct {
	ctx             context.Context
	appsCache       riov1controller.AppCache
	gitWatcherCache webhookv1controller.GitWatcherCache
	services        riov1controller.ServiceController
}

func (h Handler) updateBaseRevision(commit string, svc *riov1.Service) error {
	deepcopy := svc.DeepCopy()
	deepcopy.Spec.Build.Revision = commit
	logrus.Infof("updating revision %s to base service %s/%s", commit, svc.Namespace, svc.Name)
	if _, err := h.services.Update(deepcopy); err != nil {
		return err
	}
	return nil
}

func (h Handler) scaleDownRevisions(namespace, name string) error {
	app, err := h.appsCache.Get(namespace, name)
	if err != nil {
		return err
	}
	for _, revision := range app.Spec.Revisions {
		svc, err := h.services.Cache().Get(namespace, revision.ServiceName)
		if err != nil {
			return err
		}
		deepcopy := svc.DeepCopy()
		deepcopy.Spec.Weight = 0
		if _, err := h.services.Update(deepcopy); err != nil {
			return err
		}
		logrus.Infof("Scaling down service %s weight to 0", svc.Name)
	}
	return nil
}

func (h Handler) onChange(key string, obj *webhookv1.GitCommit) (*webhookv1.GitCommit, error) {
	if obj == nil {
		return nil, nil
	}

	gitWatcher, err := h.gitWatcherCache.Get(obj.Namespace, obj.Spec.GitWatcherName)
	if err != nil {
		return nil, err
	}

	service, err := h.services.Cache().Get(obj.Namespace, obj.Spec.GitWatcherName)
	if err != nil {
		if errors.IsNotFound(err) {
			return obj, nil
		}
		return obj, err
	}

	if obj.Spec.Commit == gitWatcher.Status.FirstCommit {
		if service.Status.FirstRevision == "" && service.Status.FirstRevision != gitWatcher.Status.FirstCommit {
			service = service.DeepCopy()
			service.Status.FirstRevision = gitWatcher.Status.FirstCommit
			_, err := h.services.Update(service)
			return obj, err
		}
		return obj, nil
	}

	return obj, webhookv1.GitWebHookExecutionConditionHandled.Once(obj, func() (runtime.Object, error) {
		appName, _ := services.AppAndVersion(service)
		specCopy := service.Spec.DeepCopy()
		specCopy.Build.Repo = obj.Spec.RepositoryURL
		specCopy.Build.Revision = obj.Spec.Commit
		specCopy.Build.Branch = ""
		specCopy.Image = ""
		specCopy.App = appName
		specCopy.Version = obj.Spec.Commit[0:5]
		if !specCopy.Build.StageOnly {
			if err := h.scaleDownRevisions(obj.Namespace, appName); err != nil {
				return obj, err
			}
			specCopy.Weight = 100
			specCopy.Rollout = true
			if specCopy.RolloutInterval == 0 {
				specCopy.RolloutInterval = 5
				specCopy.RolloutIncrement = 5
			}
		} else {
			specCopy.Weight = 0
		}
		newServiceName := name.SafeConcatName(service.Name, name.Hex(obj.Spec.RepositoryURL, 7), name.Hex(obj.Spec.Commit, 5))
		newService := riov1.NewService(service.Namespace, newServiceName, riov1.Service{
			Spec: *specCopy,
		})
		logrus.Infof("Creating new service revision, name: %s, namespace: %s, revision: %s", newService.Name, newService.Namespace, obj.Spec.Commit)
		if _, err := h.services.Create(newService); err != nil {
			return obj, err
		}
		return obj, nil
	})
}
