package gitcommit

import (
	"fmt"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func (h Handler) onChangeService(key string, obj *webhookv1.GitCommit, gitWatcher *webhookv1.GitWatcher) (*webhookv1.GitCommit, error) {
	if gitWatcher.Status.FirstCommit == "" {
		gitWatcher, err := h.gitWatcherClient.Get(gitWatcher.Namespace, gitWatcher.Name, v1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if gitWatcher.Status.FirstCommit == "" {
			return obj, fmt.Errorf("waiting for gitWatcher first commit on %s/%s", gitWatcher.Namespace, gitWatcher.Name)
		}
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
		// if git commit is from different branch do no-op
		if obj.Spec.Branch != "" && obj.Spec.Branch != service.Spec.Build.Branch {
			return obj, nil
		}

		if obj.Spec.Commit == "" {
			return obj, nil
		}

		appName, _ := services.AppAndVersion(service)
		specCopy := service.Spec.DeepCopy()
		specCopy.Build.Repo = obj.Spec.RepositoryURL
		specCopy.Build.Revision = obj.Spec.Commit
		specCopy.Build.Branch = ""
		specCopy.Image = ""
		specCopy.App = appName
		if obj.Spec.PR != "" {
			specCopy.Version = "pr-" + obj.Spec.PR
			specCopy.Build.StageOnly = true
		} else {
			specCopy.Version = "v" + obj.Spec.Commit[0:5]
		}

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
		newServiceName := serviceName(service, obj)
		newService := riov1.NewService(service.Namespace, newServiceName, riov1.Service{
			Spec: *specCopy,
		})

		if obj.Spec.PR != "" && (obj.Spec.Merged || obj.Spec.Closed) {
			logrus.Infof("PR %s is merged/closed, deleting revision, name: %s, namespace: %s, revision: %s", obj.Spec.PR, newService.Name, newService.Namespace, obj.Spec.Commit)
			if err := h.services.Delete(newService.Namespace, newService.Name, &v1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
				return obj, err
			}
			return obj, nil
		}

		logrus.Infof("Creating/Updating service revision, name: %s, namespace: %s, revision: %s", newService.Name, newService.Namespace, obj.Spec.Commit)
		if existing, err := h.services.Get(newService.Namespace, newService.Name, v1.GetOptions{}); err == nil {
			existing.Spec = newService.Spec
			existing.Status.GitCommitName = obj.Name
			if _, err := h.services.Update(existing); err != nil {
				return obj, err
			}
		} else if errors.IsNotFound(err) {
			newService.Status.GitCommitName = obj.Name
			if _, err := h.services.Create(newService); err != nil {
				return obj, err
			}
		} else {
			return obj, err
		}
		return obj, nil
	})
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

func serviceName(service *riov1.Service, obj *webhookv1.GitCommit) string {
	n := name.SafeConcatName(service.Name, name.Hex(obj.Spec.RepositoryURL, 7))
	if obj.Spec.PR != "" {
		n = name.SafeConcatName(n, "pr"+obj.Spec.PR)
	} else {
		n = name.SafeConcatName(n, name.Hex(obj.Spec.Commit, 5))
	}
	return n
}
