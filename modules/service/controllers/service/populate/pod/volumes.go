package pod

import (
	"fmt"
	"sort"

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

	sort.Strings(secrets)
	sort.Strings(configMaps)

	for _, secret := range secrets {
		result = append(result, v1.Volume{
			Name: fmt.Sprintf("secret-%s", secret),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secret,
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
