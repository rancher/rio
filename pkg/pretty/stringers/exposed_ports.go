package stringers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/mapper/mappers"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

func NewContainerPort(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &ContainerPortStringer{}
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

type ContainerPortStringer struct {
	v1.ContainerPort
}

func (e ContainerPortStringer) MaybeString() interface{} {
	buf := &strings.Builder{}
	if e.TargetPort == e.Port || e.TargetPort == 0 {
		buf.WriteString(fmt.Sprint(e.Port))
	} else {
		buf.WriteString(fmt.Sprintf("%d:%d", e.Port, e.TargetPort))
	}

	protocol := e.Protocol
	if protocol == "" {
		protocol = v1.ProtocolHTTP
	}

	if protocol != v1.ProtocolHTTP {
		buf.WriteString("/")
		buf.WriteString(strings.ToLower(string(protocol)))
	}

	if e.InternalOnly {
		buf.WriteString(",internal")
	}

	if e.HostPort {
		buf.WriteString(",hostport")
	}

	return buf.String()
}

func ParsePorts(specs ...string) (result []v1.ContainerPort, err error) {
	for _, spec := range specs {
		cp, err := parsePortBinding(spec)
		if err != nil {
			return nil, err
		}

		result = append(result, cp)
	}
	return
}

func parsePortBinding(spec string) (result v1.ContainerPort, err error) {
	parts := strings.Split(spec, ",")
	portName := ""
	opts := map[string]string{}
	if len(parts) > 1 {
		if !strings.Contains(parts[1], "=") {
			portName = parts[1]
			if len(parts) > 2 {
				opts = kv.SplitMapFromSlice(parts[2:])
			}
		} else {
			opts = kv.SplitMapFromSlice(parts[1:])
		}
	}

	ports, proto := kv.Split(parts[0], "/")
	pubStr, privStr := kv.Split(ports, ":")

	n, err := strconv.Atoi(pubStr)
	if err != nil {
		return result, errors.Wrapf(err, "failed to parse port %s", pubStr)
	}
	result.Port = int32(n)
	result.TargetPort = int32(n)
	result.Protocol = v1.Protocol(strings.ToUpper(proto))
	if portName != "" {
		result.Name = portName
	}

	if privStr != "" {
		n, err := strconv.Atoi(privStr)
		if err != nil {
			return result, errors.Wrapf(err, "failed to parse port %s", privStr)
		}
		result.TargetPort = int32(n)
	}

	if v, ok := opts["internal"]; ok && v != "false" {
		result.InternalOnly = true
	}

	if v, ok := opts["hostport"]; ok && v != "false" {
		result.HostPort = true
	}

	return
}
