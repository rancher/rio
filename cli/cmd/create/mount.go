package create

import (
	"fmt"
	"strings"

	"github.com/rancher/norman/pkg/kv"
	"github.com/rancher/rio/cli/pkg/volumespec"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func ParseMounts(spec []string) ([]riov1.Mount, error) {
	var mounts []riov1.Mount
	for _, volume := range spec {
		serviceMount, err := volumespec.ParseVolume(volume)
		if err != nil {
			return nil, err
		}

		mount := createMount(serviceMount)
		mount, err = parseAdditionalOptions(mount, volume)
		if err != nil {
			return nil, err
		}

		mounts = append(mounts, createMount(serviceMount))
	}

	return mounts, nil
}

func parseAdditionalOptions(mount riov1.Mount, spec string) (riov1.Mount, error) {
	if mount.Target == "" {
		return mount, fmt.Errorf("invalid volume spec, no target path found: %s", spec)
	}

	if !strings.Contains(mount.Source, "/") && !strings.Contains(mount.Source, "\\") && mount.VolumeOptions == nil {
		mount.VolumeOptions = &riov1.VolumeOptions{}
	}

	if len(strings.SplitN(spec, ":", 3)) < 3 {
		return mount, nil
	}

	for _, opt := range strings.Split(spec, ",") {
		key, value := kv.Split(opt, "=")
		key = strings.ToLower(key)
		switch key {
		case "driver":
			if mount.VolumeOptions == nil {
				return mount, fmt.Errorf("driver can only be used with volumes, not host bind mounts")
			}
			mount.VolumeOptions.Driver = value
		case "subpath":
			if mount.VolumeOptions == nil {
				return mount, fmt.Errorf("subpath can only be used with volumes, not host bind mounts")
			}
			mount.VolumeOptions.SubPath = value
		}
	}

	return mount, nil
}

func createMount(serviceMount volumespec.ServiceVolumeConfig) riov1.Mount {
	mount := riov1.Mount{
		Kind:     serviceMount.Type,
		ReadOnly: serviceMount.ReadOnly,
		Source:   serviceMount.Source,
		Target:   serviceMount.Target,
	}

	if serviceMount.Bind != nil {
		mount.BindOptions = &riov1.BindOptions{
			Propagation: riov1.Propagation(serviceMount.Bind.Propagation),
		}
	}

	if serviceMount.Volume != nil {
		mount.VolumeOptions = &riov1.VolumeOptions{
			NoCopy: serviceMount.Volume.NoCopy,
		}
	}

	return mount
}
