package podvolume

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/rancher/rio/features/stack/controllers/service/populate/containerlist"
	"github.com/rancher/rio/features/stack/controllers/service/populate/sidekick"
	"github.com/rancher/rio/pkg/namespace"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func NameOfVolume(volume riov1.Mount) string {
	switch volume.Kind {
	case "bind":
		return "host-" + strings.Replace(volume.Source, "/", "-", -1)
	case "volume":
		if volume.Source == "" {
			return "anon" + strings.Replace(volume.Target, "/", "-", -1)
		}
		return volume.Source
	}

	return ""
}

func NameOfConfig(config riov1.ConfigMapping) string {
	return "config-" + config.Source
}

func NameOfSecret(secret riov1.SecretMapping) string {
	return "secret-" + secret.Source
}

func UsedTemplates(volumeDefs map[string]*riov1.Volume, service *riov1.Service) map[string]*riov1.Volume {
	result := map[string]*riov1.Volume{}

	for _, m := range service.Spec.Volumes {
		if template := template(volumeDefs, m); template != nil {
			result[template.Name] = template
		}
	}

	for _, name := range sidekick.SortedKeys(service.Spec.Sidekicks) {
		for _, m := range service.Spec.Sidekicks[name].Volumes {
			if template := template(volumeDefs, m); template != nil {
				result[template.Name] = template
			}
		}
	}

	return result
}

func template(volumeDefs map[string]*riov1.Volume, mount riov1.Mount) *riov1.Volume {
	if mount.Kind != "volume" {
		return nil
	}

	volumeDef := volumeDefs[NameOfVolume(mount)]
	if volumeDef != nil && volumeDef.Spec.Template {
		return volumeDef
	}

	return nil
}

func Populate(volumeDefs map[string]*riov1.Volume, service *riov1.Service, spec *v1.PodSpec, stack *riov1.Stack) {
	volumes := map[string]v1.Volume{}

	for _, container := range containerlist.ForService(service) {
		for _, m := range container.Volumes {
			addVolumeFromMount(m, volumes, volumeDefs)
		}

		for _, s := range container.Secrets {
			addVolumeFromSecret(stack, service, s, volumes)
		}

		for _, c := range container.Configs {
			addVolumeFromConfig(stack, c, volumes)
		}
	}

	for _, name := range sortKeys(volumes) {
		spec.Volumes = append(spec.Volumes, volumes[name])
	}
}

func addVolumeFromConfig(stack *riov1.Stack, config riov1.ConfigMapping, volumes map[string]v1.Volume) {
	name := NameOfConfig(config)
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
					Name: namespace.NameRef(config.Source, stack),
				},
				DefaultMode: mode,
			},
		},
	}
}

func addVolumeFromSecret(stack *riov1.Stack, service *riov1.Service, secret riov1.SecretMapping, volumes map[string]v1.Volume) {
	t := true
	name := NameOfSecret(secret)
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
		source = "istio." + namespace.NameRef(service.Name, stack)
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
}

func addVolumeFromMount(volume riov1.Mount, volumes map[string]v1.Volume, volumeDefs map[string]*riov1.Volume) {
	name := NameOfVolume(volume)
	switch volume.Kind {
	case "bind":
		addHostBindMount(name, volumes, volume)
	case "volume":
		addVolume(name, volume, volumes, volumeDefs)
	}
}

func addVolume(name string, volume riov1.Mount, volumes map[string]v1.Volume, volumeDefs map[string]*riov1.Volume) {
	if volume.Source == "" {
		addAnonVolume(name, volume, volumes)
		return
	}

	volumeDef, ok := volumeDefs[name]
	if !ok {
		addEmptyDir(name, volume, volumes)
		return
	}

	if volumeDef.Spec.Template {
		return
	}

	addPersistentVolume(name, volumes)
}

func addAnonVolume(name string, volume riov1.Mount, volumes map[string]v1.Volume) {
	addEmptyDir(name, volume, volumes)
}

func addPersistentVolume(name string, volumes map[string]v1.Volume) {
	volumes[name] = v1.Volume{
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: name,
			},
		},
		Name: name,
	}
}

func addHostBindMount(name string, volumes map[string]v1.Volume, volume riov1.Mount) {
	volumes[name] = v1.Volume{
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: volume.Source,
			},
		},
		Name: name,
	}
}

func addEmptyDir(name string, volume riov1.Mount, volumes map[string]v1.Volume) {
	var size *resource.Quantity

	if volume.VolumeOptions != nil && volume.VolumeOptions.SizeInGB > 0 {
		q, err := resource.ParseQuantity(fmt.Sprintf("%dGi", volume.VolumeOptions.SizeInGB))
		if err == nil {
			size = &q
		}
	}

	volumes[name] = v1.Volume{
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{
				SizeLimit: size,
			},
		},
		Name: name,
	}
}

func sortKeys(m map[string]v1.Volume) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
