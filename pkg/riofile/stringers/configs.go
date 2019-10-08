package stringers

import (
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

const (
	ConfigsDefaultPath = "/run/configs"
)

type ConfigsStringer struct {
	DataMountStringer
}

func (d ConfigsStringer) MaybeString() interface{} {
	d.defaultPrefix = ConfigsDefaultPath
	return d.DataMountStringer.MaybeString()
}

func ParseConfig(config string) (v1.DataMount, error) {
	return ParseDataMount(config)
}

func ParseConfigs(configs ...string) (result []v1.DataMount, err error) {
	for _, config := range configs {
		r, err := ParseConfig(config)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return
}
