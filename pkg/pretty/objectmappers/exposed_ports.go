package objectmappers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/mapper/convert"
	"github.com/rancher/mapper/mappers"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

func NewExposedPorts(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &ExposedPortStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			objs, err := ParseExposedPorts(str)
			if err != nil {
				return nil, err
			}
			return objs[0], nil
		},
	}
}

type ExposedPortStringer struct {
	v1.ExposedPort
}

func (e ExposedPortStringer) MaybeString() interface{} {
	s := PortBindingStringer{PortBinding: e.PortBinding}.MaybeString()
	if e.Name == "" {
		return s
	}
	return convert.ToString(s) + "," + e.Name
}

func ParseExposedPorts(specs ...string) ([]riov1.ExposedPort, error) {
	var result []riov1.ExposedPort

	for _, spec := range specs {
		portSpec, name := kv.Split(spec, ",")
		portBinding, err := parsePortBinding(portSpec)
		if err != nil {
			return nil, err
		}
		result = append(result, riov1.ExposedPort{
			PortBinding: riov1.PortBinding{
				Protocol:   portBinding.Protocol,
				TargetPort: portBinding.TargetPort,
				Port:       portBinding.Port,
				IP:         portBinding.IP,
			},
			Name: name,
		})
	}

	return result, nil
}

func parsePortBinding(spec string) (riov1.PortBinding, error) {
	var (
		err                  error
		port, targetPortPair string
		result               riov1.PortBinding
	)

	parts := strings.SplitN(spec, ":", 3)
	switch len(parts) {
	case 1:
		targetPortPair = parts[0]
	case 2:
		port = parts[0]
		targetPortPair = parts[1]
	case 3:
		result.IP = parts[0]
		port = parts[1]
		targetPortPair = parts[2]
	}

	if port != "" {
		result.Port, err = strconv.ParseInt(port, 10, 0)
		if err != nil {
			return result, fmt.Errorf("invalid port number %s: %v", port, err)
		}
	}

	targetPort, proto := kv.Split(targetPortPair, "/")
	result.TargetPort, err = strconv.ParseInt(targetPort, 10, 0)
	if err != nil {
		return result, fmt.Errorf("invalid target port number %s: %v", targetPort, err)
	}

	if proto == "" {
		result.Protocol = "tcp"
	} else {
		result.Protocol = proto
	}

	return result, nil
}
