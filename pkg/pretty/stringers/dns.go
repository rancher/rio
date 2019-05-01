package stringers

import (
	"fmt"

	"github.com/rancher/mapper/mappers"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

func NewDNSOptions(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &PodDNSConfigOptionStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			objs := ParseDNSOptions(str)
			return objs[0], nil
		},
	}
}

type PodDNSConfigOptionStringer struct {
	riov1.PodDNSConfigOption
}

func (p PodDNSConfigOptionStringer) MaybeString() interface{} {
	if p.Value == nil {
		return p.Name
	}
	return fmt.Sprintf("%s:%s", p.Name, *p.Value)
}

func ParseDNSOptions(options ...string) (result []riov1.PodDNSConfigOption) {
	for _, opt := range options {
		k, v := kv.Split(opt, ":")
		podDNSOpt := riov1.PodDNSConfigOption{
			Name: k,
		}
		if v != "" {
			podDNSOpt.Value = &v
		}
		result = append(result, podDNSOpt)
	}

	return
}
