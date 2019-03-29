package container

import (
	"strings"

	"github.com/rancher/rio/modules/stack/controllers/service/populate/podvolume"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "k8s.io/api/core/v1"
)

func addVolumes(c *v1.Container, volume riov1.Mount) {
	name := podvolume.NameOfVolume(volume)
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
