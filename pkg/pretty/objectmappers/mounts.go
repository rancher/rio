package objectmappers

import (
	"fmt"
	"strings"

	"github.com/rancher/mapper/mappers"
	"github.com/rancher/rio/cli/pkg/volumespec"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

func NewMounts(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &MountStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			objs, err := ParseMounts(str)
			if err != nil {
				return nil, err
			}
			return objs[0], nil
		},
	}
}

type MountStringer struct {
	v1.Mount
}

func (m MountStringer) MaybeString() interface{} {
	result := ""
	if m.Source != "" {
		result = m.Source + ":"
	}
	result += m.Target

	opts := ""
	if m.ReadOnly {
		addOpt(opts, "ro")
	}
	if m.BindOptions != nil {
		addOpt(opts, string(m.BindOptions.Propagation))
	}
	if m.VolumeOptions != nil {
		if m.VolumeOptions.SubPath != "" {
			addOpt(opts, "subPath="+m.VolumeOptions.SubPath)
		}
		if m.VolumeOptions.Driver != "" {
			addOpt(opts, "driver="+m.VolumeOptions.Driver)
		}
	}

	if len(opts) == 0 {
		return result
	}

	return result + ":" + opts
}

func addOpt(opt, val string) string {
	if val == "" {
		return opt
	}

	if len(opt) == 0 {
		opt = ":"
	} else {
		opt += ","
	}
	return opt + val
}

func ParseMounts(spec ...string) ([]v1.Mount, error) {
	var mounts []v1.Mount
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

func parseAdditionalOptions(mount v1.Mount, spec string) (v1.Mount, error) {
	if mount.Target == "" {
		return mount, fmt.Errorf("invalid volume spec, no target path found: %s", spec)
	}

	if !strings.Contains(mount.Source, "/") && !strings.Contains(mount.Source, "\\") && mount.VolumeOptions == nil {
		mount.VolumeOptions = &v1.VolumeOptions{}
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

func createMount(serviceMount volumespec.ServiceVolumeConfig) v1.Mount {
	mount := v1.Mount{
		Kind:     serviceMount.Type,
		ReadOnly: serviceMount.ReadOnly,
		Source:   serviceMount.Source,
		Target:   serviceMount.Target,
	}

	if serviceMount.Bind != nil {
		mount.BindOptions = &v1.BindOptions{
			Propagation: v1.Propagation(serviceMount.Bind.Propagation),
		}
	}

	return mount
}
