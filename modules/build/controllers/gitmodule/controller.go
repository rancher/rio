package gitmodule

import (
	"context"
	"time"

	"github.com/rancher/rio/pkg/services"

	"github.com/sirupsen/logrus"

	"github.com/rancher/rio/modules/build/controllers/service"
	gitv1 "github.com/rancher/rio/pkg/apis/git.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	gitv1controller "github.com/rancher/rio/pkg/generated/controllers/git.rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/name"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/ticker"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	refreshInterval = 30
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := handler{
		ctx:       ctx,
		appsCache: rContext.Rio.Rio().V1().App().Cache(),
		services:  rContext.Rio.Rio().V1().Service(),
		modules:   rContext.Git.Git().V1().GitModule(),
	}
	h.start()

	updater := gitv1controller.UpdateGitModuleOnChange(rContext.Git.Git().V1().GitModule().Updater(), h.update)
	rContext.Git.Git().V1().GitModule().OnChange(ctx, "git-module-watch", updater)
	return nil
}

type handler struct {
	ctx       context.Context
	appsCache riov1controller.AppCache
	services  riov1controller.ServiceController
	modules   gitv1controller.GitModuleController
}

func (h handler) update(key string, obj *gitv1.GitModule) (*gitv1.GitModule, error) {
	if obj == nil {
		return obj, nil
	}

	commit, err := service.FirstCommit(obj.Spec.Repo, obj.Spec.Branch)
	if err != nil {
		return obj, err
	}
	if commit != obj.Status.LastRevision {
		svc, err := h.services.Cache().Get(obj.Spec.ServiceNamespace, obj.Spec.ServiceName)
		if err != nil {
			return obj, err
		}
		if obj.Status.LastRevision == "" {
			if err := h.updateBaseRevision(commit, svc); err != nil {
				return obj, err
			}
		} else {
			if err := h.createNewRevision(commit, svc, obj); err != nil {
				return obj, err
			}
		}
	}
	obj.Status.LastRevision = commit
	return obj, nil
}

func (h handler) updateBaseRevision(commit string, svc *riov1.Service) error {
	deepcopy := svc.DeepCopy()
	deepcopy.Spec.Build.Revision = commit
	logrus.Infof("updating revision %s to base service %s/%s", commit, svc.Namespace, svc.Name)
	if _, err := h.services.Update(deepcopy); err != nil {
		return err
	}
	return nil
}

func (h handler) createNewRevision(commit string, svc *riov1.Service, obj *gitv1.GitModule) error {
	appName, _ := services.AppAndVersion(svc)
	specCopy := svc.Spec.DeepCopy()
	specCopy.Build.Repo = obj.Spec.Repo
	specCopy.Build.Revision = commit
	specCopy.Build.Branch = ""
	specCopy.App = appName
	specCopy.Version = commit[0:5]
	specCopy.Image = ""
	if !specCopy.Build.StageOnly {
		if err := h.scaleDownRevisions(obj.Spec.ServiceNamespace, appName); err != nil {
			return err
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
	newServiceName := name.SafeConcatName(svc.Name, name.Hex(obj.Spec.Repo, 7), name.Hex(commit, 5))
	newService := riov1.NewService(svc.Namespace, newServiceName, riov1.Service{
		Spec: *specCopy,
	})
	logrus.Infof("Creating new service revision, name: %s, namespace: %s, revision: %s", newService.Name, newService.Namespace, commit)
	if _, err := h.services.Create(newService); err != nil {
		return err
	}
	return nil
}

func (h handler) scaleDownRevisions(namespace, name string) error {
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

func (h handler) start() {
	go func() {
		for range ticker.Context(h.ctx, refreshInterval*time.Second) {
			modules, err := h.modules.Cache().List("", labels.NewSelector())
			if err == nil {
				for _, m := range modules {
					h.modules.Enqueue(m.Namespace, m.Name)
				}
			}
		}
	}()
}
