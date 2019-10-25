package template

import (
	"context"
	"sort"
	"strings"

	"github.com/rancher/rio/pkg/services"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	riov1controller "github.com/rancher/rio/pkg/generated/controllers/rio.cattle.io/v1"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/name"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, rContext *types.Context) error {
	h := &handler{}

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
}

func (h *handler) generate(service *riov1.Service, status riov1.ServiceStatus) ([]runtime.Object, riov1.ServiceStatus, error) {
	if !service.Spec.Template || len(status.ContainerImages) == 0 {
		return nil, status, generic.ErrSkip
	}

	var images []string
	for containerName := range status.ContainerImages {
		bi := status.ContainerImages[containerName]
		if bi.ImageName == "" {
			return nil, status, generic.ErrSkip
		}
		images = append(images, bi.ImageName)
	}

	sort.Strings(images)
	name := name.Hex(strings.Join(images, "-"), 8)

	spec := service.Spec.DeepCopy()
	spec.Template = false
	setImage(service, status, spec)
	setPullSecrets(status, spec)

	return []runtime.Object{
		&riov1.Service{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: service.Namespace,
			},
			Spec: *spec,
		},
	}, status, nil
}

func setPullSecrets(status riov1.ServiceStatus, spec *riov1.ServiceSpec) {
	for _, bi := range status.ContainerImages {
		if bi.PullSecret == "" {
			continue
		}

		found := false
		for _, pi := range spec.ImagePullSecrets {
			if pi == bi.PullSecret {
				found = true
				break
			}
		}

		if !found {
			spec.ImagePullSecrets = append(spec.ImagePullSecrets, bi.PullSecret)
		}
	}
}

func setImage(service *riov1.Service, status riov1.ServiceStatus, spec *riov1.ServiceSpec) {
	spec.Image = status.ContainerImages[services.RootContainerName(service)].ImageName
	spec.ImageBuild = nil

	for i := range spec.Sidecars {
		spec.Sidecars[i].Image = status.ContainerImages[spec.Sidecars[i].Name].ImageName
		spec.Sidecars[i].ImageBuild = nil
	}
}
