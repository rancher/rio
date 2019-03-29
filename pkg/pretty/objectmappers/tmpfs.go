package objectmappers

import (
	"fmt"
	"strings"

	units "github.com/docker/go-units"
	"github.com/rancher/mapper/mappers"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

func NewTmpfs(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &TmpfsStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			objs, err := ParseTmpfs(str)
			if err != nil {
				return nil, err
			}
			return objs[0], nil
		},
	}
}

type TmpfsStringer struct {
	v1.Tmpfs
}

func (t TmpfsStringer) MaybeString() interface{} {
	opts := ""

	if t.SizeBytes == 0 {
		opts = addOpt(opts, "size="+units.BytesSize(float64(t.SizeBytes)))
	}

	if t.ReadOnly {
		opts = addOpt(opts, "ro")
	}

	return t.Path + opts
}

func ParseTmpfs(specs ...string) ([]riov1.Tmpfs, error) {
	var (
		result []riov1.Tmpfs
		err    error
	)

	for _, spec := range specs {
		var tmpfs riov1.Tmpfs

		name, opts := kv.Split(spec, ":")
		for _, opt := range strings.Split(opts, ",") {
			key, value := kv.Split(opt, "=")
			switch key {
			case "ro":
				tmpfs.ReadOnly = true
			case "size":
				tmpfs.SizeBytes, err = units.RAMInBytes(value)
				if err != nil {
					return nil, fmt.Errorf("failed to parse %s: %v", opt, err)
				}
			}
		}

		tmpfs.Path = name
		result = append(result, tmpfs)
	}

	return result, nil
}
