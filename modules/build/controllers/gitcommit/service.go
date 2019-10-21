package gitcommit

import (
	"fmt"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h Handler) onChangeService(key string, obj *webhookv1.GitCommit, gitWatcher *webhookv1.GitWatcher) (*webhookv1.GitCommit, error) {
	if obj.Spec.Commit == "" {
		return obj, nil
	}

	service, err := h.services.Cache().Get(obj.Namespace, gitWatcher.Spec.TargetServiceName)
	if err != nil {
		if errors.IsNotFound(err) {
			return obj, nil
		}
		return obj, err
	}

	service = service.DeepCopy()
	// todo: figure out how to support multiple repo watch
	if service.Spec.Template {
		// if git commit is from different branch do no-op
		if obj.Spec.Branch != "" && obj.Spec.Branch != service.Spec.ImageBuild.Branch {
			return obj, nil
		}

		if service.Spec.ImageBuild.Revision == "" {
			service.Spec.ImageBuild.Revision = obj.Spec.Commit
		} else {
			appName, _ := services.AppAndVersion(service)
			specCopy := service.Spec.DeepCopy()
			specCopy.ImageBuild.Repo = obj.Spec.RepositoryURL
			specCopy.ImageBuild.Revision = obj.Spec.Commit
			specCopy.ImageBuild.Branch = ""
			specCopy.Image = ""
			specCopy.App = appName
			if obj.Spec.PR != "" {
				specCopy.Version = "pr-" + obj.Spec.PR
				specCopy.StageOnly = true
			} else {
				specCopy.Version = "v" + obj.Spec.Commit[0:5]
			}

			if !specCopy.StageOnly {
				if err := h.scaleDownRevisions(obj.Namespace, appName); err != nil {
					return obj, err
				}
				specCopy.Weight = &[]int{100}[0]
			} else {
				specCopy.Weight = &[]int{0}[0]
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
		}
	} else {
		// if template is false, just update existing images
		containerName := gitWatcher.Spec.TargetContainerName
		update := false
		if containerName == gitWatcher.Spec.TargetServiceName {
			// use the first container
			if obj.Spec.Branch != "" && obj.Spec.Branch != service.Spec.ImageBuild.Branch {
				return obj, nil
			}
			if service.Spec.ImageBuild.Revision != obj.Spec.Commit {
				service.Spec.ImageBuild.Revision = obj.Spec.Commit
				update = true
			}
		} else {
			for i, con := range service.Spec.Sidecars {
				if con.Name == containerName {
					if service.Spec.Sidecars[i].ImageBuild.Revision != obj.Spec.Commit {
						service.Spec.Sidecars[i].ImageBuild.Revision = obj.Spec.Commit
						update = true
					}
				}
			}
		}
		if update {
			if _, err := h.services.Update(service); err != nil {
				return obj, err
			}
		}
	}

	return obj, nil

}

func (h Handler) updateBaseRevision(commit string, svc *riov1.Service) error {
	deepcopy := svc.DeepCopy()
	deepcopy.Spec.ImageBuild.Revision = commit
	logrus.Infof("updating revision %s to base service %s/%s", commit, svc.Namespace, svc.Name)
	if _, err := h.services.Update(deepcopy); err != nil {
		return err
	}
	return nil
}

func (h Handler) scaleDownRevisions(namespace, name string) error {
	revisions, err := h.services.Cache().GetByIndex(indexes.ServiceByApp, fmt.Sprintf("%s/%s", namespace, name))
	if err != nil {
		return err
	}
	for _, revision := range revisions {
		deepcopy := revision.DeepCopy()
		deepcopy.Spec.Weight = &[]int{0}[0]
		if _, err := h.services.Update(deepcopy); err != nil {
			return err
		}
		logrus.Infof("Scaling down service %s weight to 0", revision.Name)
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
