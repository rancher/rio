package stringers

import (
	"strings"

	"github.com/rancher/wrangler/pkg/kv"

	"github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func NewVolume(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &VolumeStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			return ParseVolume(str), nil
		},
	}
}

func ParseVolume(v string) v1.Volume {
	volume := v1.Volume{}
	name, path := kv.Split(v, ":")
	if path != "" {
		volume.Path = path
		volume.Name = name
	} else {
		volume.Path = name
	}
	return volume
}

type VolumeStringer struct {
	v1.Volume
}

func (v VolumeStringer) MaybeString() interface{} {
	buf := &strings.Builder{}
	if v.Name != "" {
		buf.WriteString(v.Name)
		buf.WriteString(":")
	}

	if v.Path != "" {
		buf.WriteString(v.Path)
	}
	return buf.String()
}
