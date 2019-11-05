package pod

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/rancher/rio/modules/service/controllers/service/populate/serviceports"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile/stringers"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/name"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	fieldRefs = map[string]string{
		"self/name":           "metadata.name",
		"self/namespace":      "metadata.namespace",
		"self/labels":         "metadata.labels",
		"self/annotations":    "metadata.annotations",
		"self/node":           "spec.nodeName",
		"self/serviceaccount": "spec.serviceAccountName",
		"self/hostip":         "status.hostIP",
		"self/nodeip":         "status.hostIP",
		"self/ip":             "status.podIP",
	}
	resourceRefs = map[string]string{
		"limits/cpu":                 "limits.cpu",
		"limits/memory":              "limits.memory",
		"limits/ephemeral-storage":   "limits.ephemeral-storage",
		"requests/cpu":               "requests.cpu",
		"requests/memory":            "requests.memory",
		"requests/ephemeral-storage": "requests.ephemeral-storage",
	}
	prefix = regexp.MustCompile("^v?[0-9]")
)

func containers(service *riov1.Service, init bool) (result []v1.Container) {
	for _, container := range services.ToNamedContainers(service) {
		if init != container.Init {
			continue
		}

		c := toContainer(container.Name, &container.Container)
		c.Name = container.Name
		result = append(result, c)
	}

	return
}

func toContainer(containerName string, c *riov1.Container) v1.Container {
	con := v1.Container{
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
		Env:             envs(containerName, c),
		VolumeMounts:    mounts(containerName, c),
		SecurityContext: securityContext(c),
	}

	if c.ImagePullPolicy == "" && c.Image != "" {
		if named, err := reference.ParseNormalizedNamed(c.Image); err == nil {
			if t, ok := named.(reference.Tagged); ok {
				if !prefix.MatchString(t.Tag()) {
					con.ImagePullPolicy = v1.PullAlways
				}
			} else {
				con.ImagePullPolicy = v1.PullAlways
			}
		}
	}

	if c.Image == "" {
		con.ImagePullPolicy = v1.PullIfNotPresent
	}

	return con
}

func securityContext(c *riov1.Container) *v1.SecurityContext {
	if c.ContainerSecurityContext == nil {
		return nil
	}
	return &v1.SecurityContext{
		RunAsUser:              c.RunAsUser,
		RunAsGroup:             c.RunAsGroup,
		ReadOnlyRootFilesystem: c.ReadOnlyRootFilesystem,
		Privileged:             c.Privileged,
	}
}

func mounts(containerName string, c *riov1.Container) (result []v1.VolumeMount) {
	config := dataMounts(stringers.ConfigsDefaultPath, "config", c.Configs)
	secrets := dataMounts(stringers.SecretsDefaultPath, "secret", c.Secrets)
	emptydirs := volumeMount(containerName, c.Volumes)
	return append(config, append(secrets, emptydirs...)...)
}

func dataMounts(def, name string, dataMounts []riov1.DataMount) (result []v1.VolumeMount) {
	readonly := false
	if name == "secret" {
		readonly = true
	}
	for _, config := range dataMounts {
		mount := v1.VolumeMount{
			Name: fmt.Sprintf("%s-%s", name, config.Name),
		}
		mount.MountPath = config.Target
		if mount.MountPath == "" {
			mount.MountPath = def
		}
		mount.SubPath = config.Key
		mount.ReadOnly = readonly
		result = append(result, mount)
	}

	return
}

func normalizeVolumes(containerName string, volumes []riov1.Volume) (result []riov1.Volume) {
	for i, volume := range volumes {
		if volume.Name == "" {
			if volume.Persistent {
				// name is required for persistent volumes, so ignore
				continue
			}
			volume.Name = fmt.Sprintf("%s-%d", containerName, i)
		}

		if volume.HostPath != "" {
			volume.Name = "host-" + name.Hex(volume.HostPath, 8)
		}
		result = append(result, volume)
	}
	return
}

func volumeMount(containerName string, volumes []riov1.Volume) (result []v1.VolumeMount) {
	for _, volume := range normalizeVolumes(containerName, volumes) {
		mount := v1.VolumeMount{
			Name:      fmt.Sprintf("vol-%s", volume.Name),
			MountPath: volume.Path,
		}
		result = append(result, mount)
	}
	return result
}

func envs(containerName string, c *riov1.Container) (result []v1.EnvVar) {
	for _, env := range c.Env {
		name := env.Name
		value := env.Value

		if env.ConfigMapName != "" {
			result = append(result, v1.EnvVar{
				Name: name,
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: env.ConfigMapName,
						},
						Key: env.Key,
					},
				},
			})

			continue
		}

		if env.SecretName != "" {
			result = append(result, v1.EnvVar{
				Name: name,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: env.SecretName,
						},
						Key: env.Key,
					},
				},
			})

			continue
		}

		basic := v1.EnvVar{
			Name:  name,
			Value: value,
		}

		if !strings.HasPrefix(value, "$(") || !strings.HasSuffix(value, ")") {
			result = append(result, basic)
			continue
		}

		key := strings.ToLower(value[2 : len(value)-1])

		if fieldRefValue, ok := fieldRefs[key]; ok {
			result = append(result, v1.EnvVar{
				Name: name,
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{
						FieldPath: fieldRefValue,
					},
				},
			})
			continue
		}

		if resourceRefValue, ok := resourceRefs[key]; ok {
			result = append(result, v1.EnvVar{
				Name: name,
				ValueFrom: &v1.EnvVarSource{
					ResourceFieldRef: &v1.ResourceFieldSelector{
						ContainerName: containerName,
						Resource:      resourceRefValue,
					},
				},
			})
			continue
		}
		result = append(result, basic)
	}

	return
}

func ports(c *riov1.Container) (result []v1.ContainerPort) {
	for _, port := range c.Ports {
		port = stringers.NormalizeContainerPort(port)
		if port.Port == 0 {
			continue
		}

		p := v1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.TargetPort,
			Protocol:      serviceports.Protocol(port.Protocol),
		}
		if port.HostPort {
			p.HostPort = port.Port
		}
		result = append(result, p)
	}

	return
}

func resources(c *riov1.Container) (result v1.ResourceRequirements) {
	if c.CPUMillis == nil || *c.CPUMillis == 0 {
		result.Requests = v1.ResourceList{
			v1.ResourceCPU: defaultCPU,
		}
	} else {
		q, _ := resource.ParseQuantity(fmt.Sprintf("%dm", *c.CPUMillis))
		result.Requests = v1.ResourceList{
			v1.ResourceCPU: q,
		}

	}

	if c.MemoryBytes == nil || *c.MemoryBytes == 0 {
		if result.Requests == nil {
			result.Requests = v1.ResourceList{}
		}
		result.Requests[v1.ResourceMemory] = defaultMemory
	} else {
		if result.Requests == nil {
			result.Requests = v1.ResourceList{}
		}
		q := resource.NewQuantity(*c.MemoryBytes, resource.BinarySI)
		result.Requests[v1.ResourceMemory] = *q
	}

	return
}
