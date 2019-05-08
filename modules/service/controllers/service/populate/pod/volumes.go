package pod

import (
	"fmt"
	"sort"
	"strconv"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
)

func volumes(service *riov1.Service) (result []v1.Volume) {
	secrets := secretNames(&service.Spec.Container)
	configMaps := configMapNames(&service.Spec.Container)

	for _, sidecar := range service.Spec.Sidecars {
		secrets = append(secrets, secretNames(&sidecar.Container)...)
		configMaps = append(configMaps, configMapNames(&sidecar.Container)...)
	}

	secrets = removeDuplicate(secrets)
	sort.Strings(secrets)
	configMaps = removeDuplicate(configMaps)
	sort.Strings(configMaps)

	for _, secret := range secrets {
		// todo: handle mode in api?
		var defaultMode *int32
		secretName := secret
		if secret == "identity" {
			secretName = "istio." + service.Name
		}
		result = append(result, v1.Volume{
			Name: fmt.Sprintf("secret-%s", secret),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName:  secretName,
					DefaultMode: defaultMode,
					Optional:    &[]bool{true}[0],
				},
			},
		})
	}

	for _, config := range configMaps {
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

	for i, volume := range service.Spec.Volumes {
		if volume.Name == "" {
			volume.Name = strconv.Itoa(i)
		}
		result = append(result, v1.Volume{
			Name: fmt.Sprintf("emptydir-%s", volume.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	}

	return
}

func secretNames(c *riov1.Container) (result []string) {
	for _, mount := range c.Secrets {
		result = append(result, mount.Name)
	}
	return
}

func configMapNames(c *riov1.Container) (result []string) {
	for _, mount := range c.Configs {
		result = append(result, mount.Name)
	}
	return
}

func removeDuplicate(array []string) (result []string) {
	set := map[string]struct{}{}
	for _, s := range array {
		set[s] = struct{}{}
	}
	for k := range set {
		result = append(result, k)
	}
	return result
}
