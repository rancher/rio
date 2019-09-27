package pod

import (
	"strings"

	"github.com/rancher/rio/modules/service/controllers/service/populate/rbac"
	"github.com/rancher/rio/modules/service/controllers/service/populate/servicelabels"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/stackobject"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	statsPatternAnnotationKey = "sidecar.istio.io/statsInclusionPrefixes"

	defaultEnvoyStatsMatcherInclusionPatterns = []string{
		"http",
		"cluster_manager",
		"listener_manager",
		"http_mixer_filter",
		"tcp_mixer_filter",
		"server",
		"cluster.xds-grpc",
	}
)

func Populate(service *riov1.Service, systemNamespace string, os *objectset.ObjectSet) (v1.PodTemplateSpec, error) {
	pts := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      servicelabels.ServiceLabels(service),
			Annotations: servicelabels.Merge(service.Annotations),
		},
	}

	if _, ok := pts.Annotations[statsPatternAnnotationKey]; !ok && constants.ServiceMeshMode == constants.ServiceMeshModeIstio {
		pts.Annotations[statsPatternAnnotationKey] = strings.Join(defaultEnvoyStatsMatcherInclusionPatterns, ",")
	}

	podSpec := podSpec(service, systemNamespace)
	Roles(service, &podSpec, os)
	if err := images(service, &podSpec); err != nil {
		return pts, err
	}

	PersistVolume(service, os)

	pts.Spec = podSpec
	return pts, nil
}

func Roles(service *riov1.Service, podSpec *v1.PodSpec, os *objectset.ObjectSet) {
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

func PersistVolume(service *riov1.Service, os *objectset.ObjectSet) {
	var volumes []riov1.Volume
	for _, volume := range service.Spec.Volumes {
		volumes = append(volumes, volume)
	}

	for _, c := range service.Spec.Sidecars {
		for _, volume := range c.Volumes {
			volumes = append(volumes, volume)
		}
	}

	for _, v := range volumes {
		if strings.HasPrefix(v.Name, "pv-") && constants.DefaultStorageClass {
			pv := constructors.NewPersistentVolumeClaim(service.Namespace, v.Name, v1.PersistentVolumeClaim{
				Spec: v1.PersistentVolumeClaimSpec{
					AccessModes: []v1.PersistentVolumeAccessMode{
						v1.ReadWriteOnce,
					},
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceStorage: resource.MustParse(constants.RegistryStorageSize),
						},
					},
				},
			})
			os.Add(pv)
		}
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
