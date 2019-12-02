package gitcommit

import (
	"fmt"

	webhookv1 "github.com/rancher/gitwatcher/pkg/apis/gitwatcher.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (h Handler) onChangeService(key string, obj *webhookv1.GitCommit, gitWatcher *webhookv1.GitWatcher) (*webhookv1.GitCommit, error) {
	if obj.Spec.Commit == "" {
		return obj, nil
	}

	baseService, err := h.services.Cache().Get(obj.Namespace, gitWatcher.Annotations[constants.ServiceLabel])
	if err != nil {
		if errors.IsNotFound(err) {
			return obj, nil
		}
		return obj, err
	}

	var imageBuild riov1.ImageBuildSpec
	containers := append(baseService.Spec.Sidecars, riov1.NamedContainer{
		Name:      services.RootContainerName(baseService),
		Container: baseService.Spec.Container,
	})
	containerName := gitWatcher.Annotations[constants.ContainerLabel]
	for _, container := range containers {
		if container.Name == containerName {
			imageBuild = *container.ImageBuild
			break
		}
	}

	baseService = baseService.DeepCopy()
	if baseService.Spec.Template {
		// if git commit is from different branch do no-op
		if obj.Spec.Branch != "" && obj.Spec.Branch != imageBuild.Branch {
			return obj, nil
		}

		serviceName := serviceName(baseService, obj)
		if baseService.Status.ContainerRevision == nil {
			baseService.Status.ContainerRevision = map[string]riov1.BuildRevision{}
		}
		revision := baseService.Status.ContainerRevision[containerName]
		revision.Commits = append(revision.Commits, obj.Spec.Commit)
		baseService.Status.ContainerRevision[containerName] = revision
		baseService.Status.GitCommits = append(baseService.Status.GitCommits, obj.Name)

		if obj.Spec.PR != "" && (obj.Spec.Merged || obj.Spec.Closed) {
			logrus.Infof("PR %s is merged/closed, deleting revision, name: %s, namespace: %s, revision: %s", obj.Spec.PR, serviceName, baseService.Namespace, obj.Spec.Commit)
			if baseService.Status.ShouldClean == nil {
				baseService.Status.ShouldClean = map[string]bool{}
			}
			baseService.Status.ShouldClean[serviceName] = true
		} else {
			baseService.Status.ShouldGenerate = serviceName
		}

		if _, err := h.services.UpdateStatus(baseService); err != nil {
			return obj, err
		}
	} else {
		// if template is false, just update existing images
		update := false
		if containerName == gitWatcher.Annotations[constants.ServiceLabel] {
			// use the first container
			if obj.Spec.Branch != "" && obj.Spec.Branch != baseService.Spec.ImageBuild.Branch {
				return obj, nil
			}
			if baseService.Spec.ImageBuild.Revision != obj.Spec.Commit {
				baseService.Spec.ImageBuild.Revision = obj.Spec.Commit
				update = true
			}
		} else {
			for i, con := range baseService.Spec.Sidecars {
				if con.Name == containerName {
					if baseService.Spec.Sidecars[i].ImageBuild.Revision != obj.Spec.Commit {
						baseService.Spec.Sidecars[i].ImageBuild.Revision = obj.Spec.Commit
						update = true
					}
				}
			}
		}
		if update {
			if _, err := h.services.Update(baseService); err != nil {
				return obj, err
			}
		}
	}

	return obj, nil
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
