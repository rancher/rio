package deploy

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func volumeMap(volumes []*v1beta1.Volume) map[string]*v1beta1.Volume {
	vols := map[string]*v1beta1.Volume{}

	for _, vol := range volumes {
		vols[vol.Name] = vol
	}

	return vols
}

func addHostBindMount(volumes map[string]v1.Volume, volume v1beta1.Mount) string {
	name := "host-" + strings.Replace(volume.Source, "/", "-", -1)
	volumes[name] = v1.Volume{
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: volume.Source,
			},
		},
		Name: name,
	}

	return name
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

func addAnonVolume(volume v1beta1.Mount, volumes map[string]v1.Volume) string {
	name := "anon-" + strings.Replace(volume.Target, "/", "-", -1)
	addEmptyDir(name, volume, volumes)
	return name
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

func addVolume(volume v1beta1.Mount, volumes map[string]v1.Volume, volumeDefs map[string]*v1beta1.Volume, usedTemplates map[string]*v1beta1.Volume) string {
	if volume.Source == "" {
		return addAnonVolume(volume, volumes)
	}

	name := volume.Source

	volumeDef, ok := volumeDefs[name]
	if !ok {
		addEmptyDir(name, volume, volumes)
		return name
	}

	if volumeDef.Spec.Template {
		usedTemplates[name] = volumeDef
		return name
	}

	addPersistentVolume(name, volumes)
	return name
}

func addVolumes(c *v1.Container, volume v1beta1.Mount, volumes map[string]v1.Volume, volumeDefs map[string]*v1beta1.Volume, usedTemplates map[string]*v1beta1.Volume) {
	name := ""
	switch volume.Kind {
	case "bind":
		name = addHostBindMount(volumes, volume)
	case "volume":
		name = addVolume(volume, volumes, volumeDefs, usedTemplates)
	}

	if name == "" {
		return
	}

	mount := v1.VolumeMount{
		Name:      name,
		ReadOnly:  volume.ReadOnly,
		MountPath: volume.Target,
	}

	if volume.BindOptions != nil {
		if strings.Contains(string(volume.BindOptions.Propagation), "shared") {
			prop := v1.MountPropagationBidirectional
			mount.MountPropagation = &prop
		} else if strings.Contains(string(volume.BindOptions.Propagation), "private") ||
			strings.Contains(string(volume.BindOptions.Propagation), "slave") {
			prop := v1.MountPropagationHostToContainer
			mount.MountPropagation = &prop
		}
	}

	if volume.VolumeOptions != nil {
		mount.SubPath = volume.VolumeOptions.SubPath
	}

	c.VolumeMounts = append(c.VolumeMounts, mount)
}
