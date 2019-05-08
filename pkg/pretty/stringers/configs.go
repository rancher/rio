package stringers

import (
	"github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

const (
	configsDefaultPath = "/run/configs"
)

func NewConfigs(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &ConfigsStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			objs, err := ParseConfigs(str)
			if err != nil {
				return nil, err
			}
			return objs[0], nil
		},
	}
}

type ConfigsStringer struct {
	DataMountStringer
}

func (d ConfigsStringer) MaybeString() interface{} {
	d.defaultPrefix = configsDefaultPath
	return d.DataMountStringer.MaybeString()
}

func ParseConfigs(configs ...string) ([]v1.DataMount, error) {
	return ParseDataMounts(configsDefaultPath, configs...)
}
