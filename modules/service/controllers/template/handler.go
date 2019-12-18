package template

import (
	"context"
	"fmt"

	webhookv1controller "github.com/rancher/gitwatcher/pkg/generated/controllers/gitwatcher.cattle.io/v1"
	"github.com/rancher/rio/cli/cmd/weight"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/indexes"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := &handler{
		services:   rContext.Rio.Rio().V1().Service(),
		gitcommits: rContext.Webhook.Gitwatcher().V1().GitCommit().Cache(),
	}

	riov1controller.RegisterServiceGeneratingHandler(ctx,
		rContext.Rio.Rio().V1().Service(),
		rContext.Apply.
			WithCacheTypes(rContext.Rio.Rio().V1().Service()).
			WithNoDelete(),
		"",
		"template",
		h.generate,
		nil)

	return nil
}

type handler struct {
	gitcommits webhookv1controller.GitCommitCache
	services   riov1controller.ServiceController
}

func (h *handler) generate(service *riov1.Service, status riov1.ServiceStatus) ([]runtime.Object, riov1.ServiceStatus, error) {
	skip, err := h.skip(service)
	if err != nil {
		return nil, status, err
	}
	if skip {
		return nil, status, generic.ErrSkip
	}

	if err := h.cleanup(service); err != nil {
		return nil, status, err
	}

	name := status.ShouldGenerate
	app, _ := services.AppAndVersion(service)

	spec := service.Spec.DeepCopy()
	spec.Template = false
	spec.App = app
	spec.Version = ""
	setImageBuild(service, status, spec)
	setPullSecrets(spec)

	generatedFromPR, err := h.generatedFromPR(service)
	if err != nil {
		return nil, status, err
	}
	if !service.Spec.StageOnly && !generatedFromPR {
		if err := h.scaleDownRevisions(service.Namespace, app, name); err != nil {
			return nil, status, nil
		}
		spec.Weight = &[]int{weight.PromoteWeight}[0]
	}

	if status.ShouldClean[name] || status.GeneratedServices[name] {
		return nil, status, nil
	}

	logrus.Infof("Generating service %s/%s from template", service.Namespace, name)
	return []runtime.Object{
		&riov1.Service{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: service.Namespace,
				Annotations: map[string]string{
					constants.GitCommitLabel: last(service.Status.GitCommits),
				},
			},
			Spec: *spec,
		},
	}, status, nil
}

func (h *handler) generatedFromPR(service *riov1.Service) (bool, error) {
	if len(service.Status.GitCommits) == 0 {
		return false, nil
	}

	gc, err := h.gitcommits.Get(service.Namespace, last(service.Status.GitCommits))
	if err != nil {
		return false, err
	}

	return gc.Spec.PR != "", nil
}

func (h *handler) cleanup(service *riov1.Service) error {
	for shouldDelete := range service.Status.ShouldClean {
		if err := h.services.Delete(service.Namespace, shouldDelete, &metav1.DeleteOptions{}); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (h *handler) scaleDownRevisions(namespace, name, excludedService string) error {
	revisions, err := h.services.Cache().GetByIndex(indexes.ServiceByApp, fmt.Sprintf("%s/%s", namespace, name))
	if err != nil {
		return err
	}
	for _, revision := range revisions {
		if revision.Name == excludedService {
			continue
		}
		if revision.Spec.Template {
			continue
		}
		deepcopy := revision.DeepCopy()
		if deepcopy.Spec.Weight != nil && *deepcopy.Spec.Weight == 0 {
			continue
		}
		deepcopy.Spec.Weight = &[]int{0}[0]
		if _, err := h.services.Update(deepcopy); err != nil {
			return err
		}
		logrus.Infof("Scaling down service %s weight to 0", revision.Name)
	}
	return nil
}

func (h *handler) skip(service *riov1.Service) (bool, error) {
	if service.Status.ShouldGenerate == "" {
		return true, nil
	}
	fromPR, err := h.generatedFromPR(service)
	if err != nil {
		return false, err
	}
	if fromPR {
		return false, nil
	}
	if !service.Spec.Template || len(service.Status.ContainerRevision) == 0 {
		return true, nil
	}
	needed := 0
	has := 0
	// Revision empty is a template
	if service.Spec.ImageBuild != nil && service.Spec.ImageBuild.Revision == "" {
		needed++
	}
	for _, c := range service.Spec.Sidecars {
		if c.ImageBuild != nil && c.ImageBuild.Revision == "" {
			needed++
		}
	}

	for _, c := range service.Status.ContainerRevision {
		if len(c.Commits) > 0 {
			has++
		}
	}
	return needed != has, nil
}

func setPullSecrets(spec *riov1.ServiceSpec) {
	var imagePullSecrets []string

	if spec.ImageBuild != nil && spec.ImageBuild.PushRegistrySecretName != "" {
		imagePullSecrets = append(imagePullSecrets, spec.ImageBuild.PushRegistrySecretName)
	}

	for _, con := range spec.Sidecars {
		if con.ImageBuild != nil && con.ImageBuild.PushRegistrySecretName != "" {
			imagePullSecrets = append(imagePullSecrets, con.ImageBuild.PushRegistrySecretName)
		}
	}
}

func setImageBuild(service *riov1.Service, status riov1.ServiceStatus, spec *riov1.ServiceSpec) {
	if service.Spec.ImageBuild != nil {
		spec.ImageBuild = service.Spec.ImageBuild
		spec.ImageBuild.Revision = last(status.ContainerRevision[services.RootContainerName(service)].Commits)
	}

	for i := range spec.Sidecars {
		if service.Spec.Sidecars[i].ImageBuild != nil {
			spec.Sidecars[i].ImageBuild = service.Spec.Sidecars[i].ImageBuild
			spec.Sidecars[i].ImageBuild.Revision = last(status.ContainerRevision[spec.Sidecars[i].Name].Commits)
		}
	}
}

func last(a []string) string {
	if len(a) == 0 {
		return ""
	}
	return a[len(a)-1]
}
