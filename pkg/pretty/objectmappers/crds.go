package objectmappers

import (
	"fmt"

	"github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

func NewCRDs(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &v1.CustomResourceDefinition{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			return ParseCustomResourceDefintion(str)[0], nil
		},
	}
}

type CustomResourceDefinitionStringer struct {
	v1.CustomResourceDefinition
}

func (c *CustomResourceDefinitionStringer) MaybeString() interface{} {
	return fmt.Sprintf("%s.%s/%s", c.Kind, c.Group, c.Version)
}

func ParseCustomResourceDefintion(crds ...string) (ret []v1.CustomResourceDefinition) {
	for _, crd := range crds {
		nameGroup, version := kv.Split(crd, "/")
		name, group := kv.Split(nameGroup, ".")
		ret = append(ret, v1.CustomResourceDefinition{
			Kind:    name,
			Group:   group,
			Version: version,
		})
	}

	return
}
