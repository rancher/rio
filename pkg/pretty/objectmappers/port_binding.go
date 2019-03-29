package objectmappers

import (
	"bytes"
	"strconv"

	"github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

func NewPortBinding(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &PortBindingStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			objs, err := ParsePorts(str)
			if err != nil {
				return nil, err
			}
			return objs[0], nil
		},
	}
}

type PortBindingStringer struct {
	v1.PortBinding
}

func (p PortBindingStringer) MaybeString() interface{} {
	b := bytes.Buffer{}
	if p.Port != 0 && p.TargetPort != 0 {
		if p.IP != "" {
			b.WriteString(p.IP)
			b.WriteString(":")
		}
		b.WriteString(strconv.FormatInt(p.Port, 10))
		b.WriteString(":")
		b.WriteString(strconv.FormatInt(p.TargetPort, 10))
	} else if p.TargetPort != 0 {
		b.WriteString(strconv.FormatInt(p.TargetPort, 10))
	}

	if b.Len() > 0 && p.Protocol != "" && p.Protocol != "tcp" {
		b.WriteString("/")
		b.WriteString(p.Protocol)
	}

	return b.String()
}

func ParsePorts(specs ...string) ([]v1.PortBinding, error) {
	var result []v1.PortBinding

	for _, spec := range specs {
		portBinding, err := parsePortBinding(spec)
		if err != nil {
			return nil, err
		}
		result = append(result, portBinding)
	}

	return result, nil
}
