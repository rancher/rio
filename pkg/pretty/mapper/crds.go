package mapper

import (
	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
)

func NewCRDs(field string) ObjectsToSlice {
	return ObjectsToSlice{
		Field: field,
		NewObject: func() MaybeStringer {
			return &v1beta1.CustomResourceDefinition{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			nameGroup, version := kv.Split(str, "/")
			name, group := kv.Split(nameGroup, ".")
			return map[string]interface{}{
				"kind":    name,
				"group":   group,
				"version": version,
			}, nil
		},
	}
}
