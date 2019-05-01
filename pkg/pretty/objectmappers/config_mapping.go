package objectmappers

import (
	"fmt"

	"github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func NewConfigMapping(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &ConfigMappingStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			objs, err := ParseConfigMapping(str)
			if err != nil {
				return nil, err
			}
			return objs[0], nil
		},
	}
}

type ConfigMappingStringer struct {
	v1.ConfigMapping
}

func (c ConfigMappingStringer) MaybeString() interface{} {
	if c.Target == "/"+c.Source {
		c.Target = ""
	}

	msg := c.Source
	if c.Target != "" {
		msg += ":" + c.Target
	}

	if c.UID > 0 {
		msg = fmt.Sprintf("%s,uid=%d", msg, c.UID)
	}

	if c.GID > 0 {
		msg = fmt.Sprintf("%s,gid=%d", msg, c.GID)
	}

	if c.Mode != "" {
		msg = fmt.Sprintf("%s,mode=%s", msg, c.Mode)
	}

	return msg
}
