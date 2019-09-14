package pod

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/rancher/rio/modules/service/controllers/service/populate/serviceports"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
)

var (
	fieldRefs = map[string]string{
		"self/name":           "metadata.name",
		"self/namespace":      "metadata.namespace",
		"self/labels":         "metadata.labels",
		"self/annotations":    "metadata.annotations",
		"self/node":           "spec.nodeName",
		"self/serviceAccount": "spec.serviceAccountName",
		"self/hostIp":         "status.hostIP",
		"self/nodeIp":         "status.hostIP",
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
)

func containers(service *riov1.Service, systemNamespace string, init bool) (result []v1.Container) {
	system := service.Namespace == systemNamespace
	if !init && !reflect.DeepEqual(service.Spec.Container, riov1.Container{}) {
		c := toContainer(service.Name, &service.Spec.Container, system)
		c.Name = service.Name
		result = append(result, c)
	}

	for _, sidecar := range service.Spec.Sidecars {
		if sidecar.Init != init {
			continue
		}

		c := toContainer(sidecar.Name, &sidecar.Container, system)
		c.Name = sidecar.Name
		result = append(result, c)
	}

	return
}

func toContainer(containerName string, c *riov1.Container, system bool) v1.Container {
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
		Ports:           ports(c, system),
		Env:             envs(containerName, c),
		VolumeMounts:    mounts(c),
		SecurityContext: securityContext(c),
	}

	if system && c.SecurityContext != nil {
		con.SecurityContext = c.SecurityContext
	}
	return con
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
	secrets := dataMounts("secret", c.Secrets)
	emptydirs := volumeMount("emptydir", c.Volumes)
	return append(config, append(secrets, emptydirs...)...)
}

func dataMounts(name string, dataMounts []riov1.DataMount) (result []v1.VolumeMount) {
	readonly := false
	if name == "secret" {
		readonly = true
	}
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
				mount.MountPath = filepath.Join(config.Directory, config.File)
			}
			mount.SubPath = config.Key
		}
		mount.ReadOnly = readonly
		result = append(result, mount)
	}

	return
}

func volumeMount(name string, volumes []riov1.Volume) (result []v1.VolumeMount) {
	for i, volume := range volumes {
		if volume.Name == "" {
			volume.Name = strconv.Itoa(i)
		}
		mount := v1.VolumeMount{
			Name:      fmt.Sprintf("%s-%s", name, volume.Name),
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

		key := value[2 : len(value)-1]

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

func ports(c *riov1.Container, system bool) (result []v1.ContainerPort) {
	for _, port := range c.Ports {
		p := v1.ContainerPort{
			Name:          port.Name,
			ContainerPort: port.TargetPort,
			Protocol:      serviceports.Protocol(port.Protocol),
		}
		if system && port.HostPort {
			p.HostPort = port.Port
		}
		result = append(result, p)
	}

	return
}

func resources(c *riov1.Container) (result v1.ResourceRequirements) {
	if c.CPUs == nil || c.CPUs.IsZero() {
		result.Requests = v1.ResourceList{
			v1.ResourceCPU: defaultCPU,
		}
	} else {
		result.Requests = v1.ResourceList{
			v1.ResourceCPU: *c.CPUs,
		}

	}

	if c.Memory == nil || c.Memory.IsZero() {
		if result.Requests == nil {
			result.Requests = v1.ResourceList{}
		}
		result.Requests[v1.ResourceMemory] = defaultMemory
	} else {
		if result.Requests == nil {
			result.Requests = v1.ResourceList{}
		}
		result.Requests[v1.ResourceMemory] = *c.Memory
	}

	return
}
