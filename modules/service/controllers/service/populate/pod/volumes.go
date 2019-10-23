package pod

import (
	"fmt"
	"sort"

	"github.com/rancher/rio/pkg/services"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
)

func secretVolumes(containers []riov1.NamedContainer) (result []v1.Volume) {
	var names []string
	for _, container := range containers {
		for _, mount := range container.Secrets {
			if mount.Name != "" {
				names = append(names, mount.Name)
			}
		}
	}

	for _, secret := range removeDuplicateAndSort(names) {
		result = append(result, v1.Volume{
			Name: fmt.Sprintf("secret-%s", secret),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secret,
					Optional:   &[]bool{true}[0],
				},
			},
		})
	}

	return
}

func configVolumes(containers []riov1.NamedContainer) (result []v1.Volume) {
	var names []string
	for _, container := range containers {
		for _, mount := range container.Configs {
			if mount.Name != "" {
				names = append(names, mount.Name)
			}
		}
	}

	for _, config := range removeDuplicateAndSort(names) {
		result = append(result, v1.Volume{
			Name: fmt.Sprintf("config-%s", config),
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: config,
					},
				},
			},
		})
	}

	return
}

type sortedVolumes struct {
	All      map[string]bool
	EmptyDir map[string]riov1.Volume
	HostPath map[string]riov1.Volume
	PVC      map[string]riov1.Volume
}

func sortVolumes(containers []riov1.NamedContainer, volumeTemplates map[string]riov1.VolumeTemplate) (result sortedVolumes) {
	result.All = map[string]bool{}
	result.EmptyDir = map[string]riov1.Volume{}
	result.HostPath = map[string]riov1.Volume{}
	result.PVC = map[string]riov1.Volume{}

	for _, container := range containers {
		for _, volume := range normalizeVolumes(container.Name, container.Volumes) {
			if result.All[volume.Name] {
				continue
			}
			result.All[volume.Name] = true
			if volume.HostPath != "" {
				result.HostPath[volume.Name] = volume
			} else if _, ok := volumeTemplates[volume.Name]; !ok && volume.Persistent {
				result.PVC[volume.Name] = volume
			} else if !volume.Persistent {
				result.EmptyDir[volume.Name] = volume
			}
		}
	}

	return
}

func emptyDirVolumes(emptyDir map[string]riov1.Volume) (result []v1.Volume) {
	var names []string
	for name := range emptyDir {
		names = append(names, name)
	}
	for _, name := range removeDuplicateAndSort(names) {
		result = append(result, v1.Volume{
			Name: fmt.Sprintf("vol-%s", name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	}
	return
}

func hostPathVolumes(hostPath map[string]riov1.Volume) (result []v1.Volume) {
	var names []string
	for name := range hostPath {
		names = append(names, name)
	}
	for _, name := range removeDuplicateAndSort(names) {
		result = append(result, v1.Volume{
			Name: fmt.Sprintf("vol-%s", name),
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: hostPath[name].HostPath,
					Type: hostPath[name].HostPathType,
				},
			},
		})
	}
	return
}

func pvcVolumes(pvcs map[string]riov1.Volume) (result []v1.Volume) {
	var names []string
	for name := range pvcs {
		names = append(names, name)
	}
	for _, name := range removeDuplicateAndSort(names) {
		result = append(result, v1.Volume{
			Name: fmt.Sprintf("vol-%s", name),
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: name,
				},
			},
		})
	}
	return
}

func NormalizeVolumeTemplates(service *riov1.Service) map[string]riov1.VolumeTemplate {
	templates := map[string]riov1.VolumeTemplate{}
	for _, template := range service.Spec.VolumeTemplates {
		if _, ok := templates[template.Name]; ok || template.Name == "" {
			continue
		}
		templates[template.Name] = template
	}
	return templates
}

func diskVolumes(containers []riov1.NamedContainer, service *riov1.Service) (result []v1.Volume) {
	templates := NormalizeVolumeTemplates(service)
	sorted := sortVolumes(containers, templates)

	result = append(result, emptyDirVolumes(sorted.EmptyDir)...)
	result = append(result, hostPathVolumes(sorted.HostPath)...)
	result = append(result, pvcVolumes(sorted.PVC)...)
	return
}

func volumes(service *riov1.Service) (result []v1.Volume) {
	containers := services.ToNamedContainers(service)
	result = append(result, secretVolumes(containers)...)
	result = append(result, configVolumes(containers)...)
	result = append(result, diskVolumes(containers, service)...)
	return
}

func removeDuplicateAndSort(array []string) (result []string) {
	set := map[string]struct{}{}
	for _, s := range array {
		set[s] = struct{}{}
	}
	for k := range set {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
