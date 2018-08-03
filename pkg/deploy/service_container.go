package deploy

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func container(serviceName, name string, container v1beta1.ContainerConfig, volumes map[string]v1.Volume, volumeDefs map[string]*v1beta1.Volume, usedTemplates map[string]*v1beta1.Volume) v1.Container {
	c := v1.Container{
		Name:            name,
		Image:           container.Image,
		Command:         container.Entrypoint,
		Args:            container.Command,
		WorkingDir:      container.WorkingDir,
		ImagePullPolicy: v1.PullIfNotPresent,
		SecurityContext: &v1.SecurityContext{
			ReadOnlyRootFilesystem: &container.ReadonlyRootfs,
			Capabilities: &v1.Capabilities{
				Add:  toCaps(container.CapAdd),
				Drop: toCaps(container.CapDrop),
			},
			Privileged: &container.Privileged,
		},
		TTY:   container.Tty,
		Stdin: container.OpenStdin,
		Resources: v1.ResourceRequirements{
			Limits:   v1.ResourceList{},
			Requests: v1.ResourceList{},
		},
	}

	switch container.ImagePullPolicy {
	case "never":
		c.ImagePullPolicy = v1.PullNever
	case "always":
		c.ImagePullPolicy = v1.PullAlways
	}

	populateResources(&c, container)

	if n, err := strconv.ParseInt(container.User, 10, 0); err == nil {
		c.SecurityContext.RunAsUser = &n
	}

	populateEnv(&c, container)

	c.LivenessProbe, c.ReadinessProbe = toProbes(container.Healthcheck)

	for _, volume := range container.Volumes {
		addVolumes(&c, volume, volumes, volumeDefs, usedTemplates)
	}

	addConfigs(&c, container, volumes)
	addSecrets(serviceName, &c, container, volumes)

	return c
}

func addConfigs(c *v1.Container, container v1beta1.ContainerConfig, volumes map[string]v1.Volume) {
	for _, config := range container.Configs {
		name := "config-" + config.Source
		var mode *int32
		if config.Mode != "" {
			r, err := strconv.ParseInt(config.Mode, 8, 32)
			if err == nil {
				r32 := int32(r)
				mode = &r32
			}
		}

		volumes[name] = v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: config.Source,
					},
					DefaultMode: mode,
				},
			},
		}

		c.VolumeMounts = append(c.VolumeMounts, v1.VolumeMount{
			Name:      name,
			MountPath: config.Target,
			SubPath:   "content",
		})
	}
}

func populateEnv(c *v1.Container, container v1beta1.ContainerConfig) {
	for _, env := range container.Environment {
		name, value := kv.Split(env, "=")
		c.Env = append(c.Env, toEnvVar(c.Name, name, value))
	}
}

func toEnvVar(containerName, name, value string) v1.EnvVar {
	basic := v1.EnvVar{
		Name:  name,
		Value: value,
	}

	if !strings.HasPrefix(value, "$(") || !strings.HasSuffix(value, ")") {
		return basic
	}

	key := value[2 : len(value)-1]

	if fieldRefValue, ok := fieldRefs[key]; ok {
		return v1.EnvVar{
			Name: name,
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: fieldRefValue,
				},
			},
		}
	}

	if resourceRefValue, ok := resourceRefs[key]; ok {
		return v1.EnvVar{
			Name: name,
			ValueFrom: &v1.EnvVarSource{
				ResourceFieldRef: &v1.ResourceFieldSelector{
					ContainerName: containerName,
					Resource:      resourceRefValue,
				},
			},
		}
	}

	k, v := kv.Split(key, "/")
	optional := strings.HasSuffix(v, "?")
	if optional {
		v = v[:len(v)-1]
	}

	if v == "" {
		return basic
	}

	switch k {
	case "config":
		return v1.EnvVar{
			Name: name,
			ValueFrom: &v1.EnvVarSource{
				ConfigMapKeyRef: &v1.ConfigMapKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: v,
					},
					Key:      "content",
					Optional: &optional,
				},
			},
		}
	case "secret":
		return v1.EnvVar{
			Name: name,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: v,
					},
					Key:      "content",
					Optional: &optional,
				},
			},
		}
	default:
		if resourceRefValue, ok := resourceRefs[v]; ok {
			return v1.EnvVar{
				Name: name,
				ValueFrom: &v1.EnvVarSource{
					ResourceFieldRef: &v1.ResourceFieldSelector{
						ContainerName: k,
						Resource:      resourceRefValue,
					},
				},
			}
		}
	}

	return basic
}

func populateResources(c *v1.Container, container v1beta1.ContainerConfig) {
	if container.MemoryLimitBytes > 0 {
		c.Resources.Limits[v1.ResourceMemory] = *resource.NewQuantity(container.MemoryLimitBytes, resource.DecimalSI)
	}

	if container.MemoryReservationBytes > 0 {
		c.Resources.Requests[v1.ResourceMemory] = *resource.NewQuantity(container.MemoryReservationBytes, resource.DecimalSI)
	}

	if container.CPUs != "" {
		q, err := resource.ParseQuantity(container.CPUs)
		if err == nil {
			c.Resources.Requests[v1.ResourceCPU] = q
		}
		logrus.Errorf("Failed to parse CPU request: %v", err)
	}
}

func toProbes(healthcheck *v1beta1.HealthConfig) (*v1.Probe, *v1.Probe) {
	if healthcheck == nil {
		return nil, nil
	}

	probe := v1.Probe{
		InitialDelaySeconds: int32(healthcheck.InitialDelaySeconds),
		TimeoutSeconds:      int32(healthcheck.TimeoutSeconds),
		PeriodSeconds:       int32(healthcheck.IntervalSeconds),
		SuccessThreshold:    int32(healthcheck.HealthyThreshold),
		FailureThreshold:    int32(healthcheck.UnhealthyThreshold),
	}

	test := healthcheck.Test[0]
	if strings.HasPrefix(test, "http://") || strings.HasPrefix(test, "https://") {
		u, err := url.Parse(test)
		if err == nil {
			probe.HTTPGet = &v1.HTTPGetAction{
				Path: u.Path,
			}
			if strings.HasPrefix(test, "http://") {
				probe.HTTPGet.Scheme = v1.URISchemeHTTP
			} else if strings.HasPrefix(test, "https://") {
				probe.HTTPGet.Scheme = v1.URISchemeHTTPS
			}

			port := u.Port()
			if port == "" && probe.HTTPGet.Scheme == v1.URISchemeHTTPS {
				probe.HTTPGet.Port = intstr.Parse("443")
			} else if port == "" {
				probe.HTTPGet.Port = intstr.Parse("80")
			} else {
				probe.HTTPGet.Port = intstr.Parse(u.Port())
			}

			for i := 1; i < len(healthcheck.Test); i++ {
				name, value := kv.Split(healthcheck.Test[i], "=")
				probe.HTTPGet.HTTPHeaders = append(probe.HTTPGet.HTTPHeaders, v1.HTTPHeader{
					Name:  name,
					Value: value,
				})
			}
		}
	} else if strings.HasPrefix(test, "tcp://") {
		u, err := url.Parse(test)
		if err == nil {
			probe.TCPSocket = &v1.TCPSocketAction{
				Port: intstr.Parse(u.Port()),
			}
		}
	} else if test == "CMD" {
		probe.Exec = &v1.ExecAction{
			Command: healthcheck.Test[1:],
		}
	} else if test == "CMD-SHELL" {
		if len(healthcheck.Test) == 2 {
			probe.Exec = &v1.ExecAction{
				Command: []string{"sh", "-c", healthcheck.Test[1]},
			}
		}
	} else if test == "NONE" {
		return nil, nil
	} else {
		probe.Exec = &v1.ExecAction{
			Command: healthcheck.Test,
		}
	}

	liveness := probe
	liveness.SuccessThreshold = 1
	return &liveness, &probe
}

func toCaps(args []string) []v1.Capability {
	var caps []v1.Capability
	for _, arg := range args {
		caps = append(caps, v1.Capability(arg))
	}
	return caps
}

func addSecrets(serviceName string, c *v1.Container, container v1beta1.ContainerConfig, volumes map[string]v1.Volume) {
	t := true

	for _, secret := range container.Secrets {
		name := "secret-" + secret.Source
		var mode *int32
		if secret.Mode != "" {
			r, err := strconv.ParseInt(secret.Mode, 8, 32)
			if err == nil {
				r32 := int32(r)
				mode = &r32
			}
		}

		source := secret.Source
		if source == "identity" {
			source = "istio." + serviceName
		}
		volumes[name] = v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					Optional:    &t,
					SecretName:  source,
					DefaultMode: mode,
				},
			},
		}

		target := secret.Target
		if target == "" {
			target = "/run/secrets/" + secret.Source
		}

		c.VolumeMounts = append(c.VolumeMounts, v1.VolumeMount{
			Name:      name,
			MountPath: target,
		})
	}
}
