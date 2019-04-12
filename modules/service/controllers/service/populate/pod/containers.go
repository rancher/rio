package pod

import (
	"fmt"
	"path/filepath"

	"github.com/rancher/rio/modules/service/controllers/service/populate/serviceports"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
)

func containers(service *riov1.Service, init bool) (result []v1.Container) {
	if !init {
		c := toContainer(&service.Spec.Container)
		c.Name = service.Name
		result = append(result, c)
	}

	for _, sidecar := range service.Spec.Sidecars {
		if sidecar.Init != init {
			continue
		}

		c := toContainer(&sidecar.Container)
		c.Name = sidecar.Name
		result = append(result, c)
	}

	return
}

func toContainer(c *riov1.Container) v1.Container {
	return v1.Container{
		Image:           c.Image,
		Command:         c.Command,
		Args:            c.Args,
		WorkingDir:      c.WorkingDir,
		LivenessProbe:   c.LivenessProbe,
		ReadinessProbe:  c.ReadinessProbe,
		ImagePullPolicy: c.ImagePullPolicy,
		Stdin:           c.Stdin,
		StdinOnce:       c.StdinOnce,
		TTY:             c.TTY,
		Resources:       resources(c),
		Ports:           ports(c),
		Env:             envs(c),
		VolumeMounts:    mounts(c),
		SecurityContext: securityContext(c),
	}
}

func securityContext(c *riov1.Container) *v1.SecurityContext {
	if c.RunAsUser != nil ||
		c.RunAsGroup != nil ||
		c.ReadOnlyRootFilesystem != nil {
		return &v1.SecurityContext{
			RunAsUser:              c.RunAsUser,
			RunAsGroup:             c.RunAsGroup,
			ReadOnlyRootFilesystem: c.ReadOnlyRootFilesystem,
		}
	}
	return nil
}

func mounts(c *riov1.Container) (result []v1.VolumeMount) {
	config := dataMounts("config", c.Configs)
	secrets := dataMounts("secrets", c.Secrets)
	return append(config, secrets...)
}

func dataMounts(name string, dataMounts []riov1.DataMount) (result []v1.VolumeMount) {
	for _, config := range dataMounts {
		mount := v1.VolumeMount{
			Name: fmt.Sprintf("%s-%s", name, config.Name),
		}
		if config.Key == "" {
			mount.MountPath = config.Directory
		} else {
			if config.File == "" {
				mount.MountPath = filepath.Join(config.Directory, config.Key)
			} else {
				mount.MountPath = config.Directory
			}
			mount.SubPath = config.Key
		}
		result = append(result, mount)
	}

	return
}

func envs(c *riov1.Container) (result []v1.EnvVar) {
	for _, env := range c.Env {
		envVar := v1.EnvVar{
			Name:  env.Name,
			Value: env.Value,
		}

		if env.ConfigMapName != "" {
			envVar.ValueFrom = &v1.EnvVarSource{
				ConfigMapKeyRef: &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: env.ConfigMapName,
					},
					Key:      env.Key,
					Optional: env.Optional,
				},
			}
		} else if env.SecretName != "" {
			envVar.ValueFrom = &v1.EnvVarSource{
				ConfigMapKeyRef: &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: env.ConfigMapName,
					},
					Key:      env.Key,
					Optional: env.Optional,
				},
			}
		}

		result = append(result, envVar)
	}

	return
}

func ports(c *riov1.Container) (result []v1.ContainerPort) {
	for _, port := range c.Ports {
		result = append(result, v1.ContainerPort{
			ContainerPort: port.TargetPort,
			Protocol:      serviceports.Protocol(port.Protocol),
		})
	}

	return
}

func resources(c *riov1.Container) (result v1.ResourceRequirements) {
	if c.CPUs.IsZero() {
		result.Requests = v1.ResourceList{
			v1.ResourceCPU: defaultCPU,
		}
	} else {
		result.Requests = v1.ResourceList{
			v1.ResourceCPU: c.CPUs,
		}

	}

	if c.Memory.IsZero() {
		if result.Requests == nil {
			result.Requests = v1.ResourceList{}
		}
		result.Requests[v1.ResourceMemory] = defaultMemory
	} else {
		if result.Requests == nil {
			result.Requests = v1.ResourceList{}
		}
		result.Requests[v1.ResourceMemory] = c.Memory
	}

	return
}
