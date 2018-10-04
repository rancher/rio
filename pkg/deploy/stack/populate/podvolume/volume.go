package podvolume

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/rancher/rio/pkg/deploy/stack/populate/containerlist"

	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/populate/sidekick"

	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func NameOfVolume(volume v1beta1.Mount) string {
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

func NameOfConfig(config v1beta1.ConfigMapping) string {
	return "config-" + config.Source
}

func NameOfSecret(secret v1beta1.SecretMapping) string {
	return "secret-" + secret.Source
}

func UsedTemplates(stack *input.Stack, service *v1beta1.Service) map[string]*v1beta1.Volume {
	result := map[string]*v1beta1.Volume{}
	volumeDefs := volumesDefsByName(stack.Volumes)

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

func template(volumeDefs map[string]*v1beta1.Volume, mount v1beta1.Mount) *v1beta1.Volume {
	if mount.Kind != "volume" {
		return nil
	}

	volumeDef := volumeDefs[NameOfVolume(mount)]
	if volumeDef != nil && volumeDef.Spec.Template {
		return volumeDef
	}

	return nil
}

func Populate(stack *input.Stack, service *v1beta1.Service, spec *v1.PodSpec) {
	volumes := map[string]v1.Volume{}
	volumeDefs := volumesDefsByName(stack.Volumes)

	for _, container := range containerlist.ForService(service) {
		for _, m := range container.Volumes {
			addVolumeFromMount(m, volumes, volumeDefs)
		}

		for _, s := range container.Secrets {
			addVolumeFromSecret(service, s, volumes)
		}

		for _, c := range container.Configs {
			addVolumeFromConfig(c, volumes)
		}
	}

	for _, name := range sortKeys(volumes) {
		spec.Volumes = append(spec.Volumes, volumes[name])
	}
}

func addVolumeFromConfig(config v1beta1.ConfigMapping, volumes map[string]v1.Volume) {
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
					Name: config.Source,
				},
				DefaultMode: mode,
			},
		},
	}
}

func addVolumeFromSecret(service *v1beta1.Service, secret v1beta1.SecretMapping, volumes map[string]v1.Volume) {
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
		source = "istio." + service.Name
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

func addVolumeFromMount(volume v1beta1.Mount, volumes map[string]v1.Volume, volumeDefs map[string]*v1beta1.Volume) {
	name := NameOfVolume(volume)
	switch volume.Kind {
	case "bind":
		addHostBindMount(name, volumes, volume)
	case "volume":
		addVolume(name, volume, volumes, volumeDefs)
	}
}

func volumesDefsByName(volumeDefs []*v1beta1.Volume) map[string]*v1beta1.Volume {
	result := map[string]*v1beta1.Volume{}
	for _, vd := range volumeDefs {
		result[vd.Name] = vd
	}
	return result

}

func addVolume(name string, volume v1beta1.Mount, volumes map[string]v1.Volume, volumeDefs map[string]*v1beta1.Volume) {
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

func addAnonVolume(name string, volume v1beta1.Mount, volumes map[string]v1.Volume) {
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

func addHostBindMount(name string, volumes map[string]v1.Volume, volume v1beta1.Mount) {
	volumes[name] = v1.Volume{
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: volume.Source,
			},
		},
		Name: name,
	}
}

func addEmptyDir(name string, volume v1beta1.Mount, volumes map[string]v1.Volume) {
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
