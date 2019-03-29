package objectmappers

import (
	"fmt"
	"strconv"

	"github.com/rancher/mapper/mappers"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
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

func ParseConfigMapping(configs ...string) ([]riov1.ConfigMapping, error) {
	var result []riov1.ConfigMapping
	for _, config := range configs {
		mapping, err := parseConfig(config)
		if err != nil {
			return nil, err
		}
		result = append(result, mapping)
	}

	return result, nil
}

func parseConfig(device string) (riov1.ConfigMapping, error) {
	result := riov1.ConfigMapping{}

	mapping, optStr := kv.Split(device, ",")
	result.Source, result.Target = kv.Split(mapping, ":")
	opts := kv.SplitMap(optStr, ",")

	if i, err := strconv.Atoi(opts["uid"]); err == nil {
		result.UID = i
	}

	if i, err := strconv.Atoi(opts["gid"]); err == nil {
		result.GID = i
	}
	result.Mode = opts["mode"]

	return result, nil
}
