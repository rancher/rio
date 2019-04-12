package pod

import (
	"github.com/rancher/rio/modules/service/controllers/service/populate/rbac"
	"github.com/rancher/rio/modules/service/controllers/service/populate/servicelabels"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Populate(service *riov1.Service, os *objectset.ObjectSet) (v1.PodTemplateSpec, error) {
	pts := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      servicelabels.ServiceLabels(service),
			Annotations: servicelabels.Merge(service.Annotations),
		},
	}

	podSpec := podSpec(service)
	roles(service, &podSpec, os)
	if err := images(service, &podSpec); err != nil {
		return pts, err
	}

	pts.Spec = podSpec
	return pts, nil
}

func roles(service *riov1.Service, podSpec *v1.PodSpec, os *objectset.ObjectSet) {
	if err := rbac.Populate(service, os); err != nil {
		os.AddErr(err)
		return
	}

	serviceAccountName := rbac.ServiceAccountName(service)
	if serviceAccountName != "" {
		podSpec.ServiceAccountName = serviceAccountName
		podSpec.AutomountServiceAccountToken = nil
	}
}

func images(service *riov1.Service, podSpec *v1.PodSpec) error {
	for i, container := range podSpec.InitContainers {
		image := service.Status.ContainerImages[container.Name]
		if image != "" {
			podSpec.InitContainers[i].Image = image
		}
		if podSpec.InitContainers[i].Image == "" {
			return stackobject.ErrSkipObjectSet
		}
	}

	for i, container := range podSpec.Containers {
		image := service.Status.ContainerImages[container.Name]
		if image != "" {
			podSpec.Containers[i].Image = image
		}
		if podSpec.Containers[i].Image == "" {
			return stackobject.ErrSkipObjectSet
		}
	}

	return nil
}
