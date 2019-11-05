package stringers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
)

var (
	keywords = map[string]bool{
		"hostport": true,
		"expose":   true,
	}
)

type ContainerPortStringer struct {
	v1.ContainerPort
}

func normalizeProtocol(proto string) v1.Protocol {
	protocol := v1.Protocol(strings.ToUpper(proto))
	switch protocol {
	case v1.ProtocolTCP:
		return v1.ProtocolTCP
	case v1.ProtocolUDP:
		return v1.ProtocolUDP
	case v1.ProtocolHTTP2:
		return v1.ProtocolHTTP2
	case v1.ProtocolSCTP:
		return v1.ProtocolSCTP
	case v1.ProtocolGRPC:
		return v1.ProtocolGRPC
	case v1.ProtocolHTTP:
		return v1.ProtocolHTTP
	default:
		return ""
	}
}

func (e ContainerPortStringer) MaybeString() interface{} {
	e.ContainerPort = NormalizeContainerPort(e.ContainerPort)

	buf := &strings.Builder{}

	port := e.Port
	if port <= 0 {
		port = 80
	}

	buf.WriteString(fmt.Sprint(port))
	if e.TargetPort > 0 && e.TargetPort != port {
		buf.WriteString(fmt.Sprintf(":%d", e.TargetPort))
	}

	protocol := normalizeProtocol(string(e.Protocol))
	switch protocol {
	case v1.ProtocolTCP:
		fallthrough
	case v1.ProtocolUDP:
		fallthrough
	case v1.ProtocolHTTP2:
		fallthrough
	case v1.ProtocolSCTP:
		fallthrough
	case v1.ProtocolGRPC:
		buf.WriteString("/")
		buf.WriteString(strings.ToLower(string(protocol)))
	case v1.ProtocolHTTP:
	default:
	}

	if e.Expose != nil {
		if *e.Expose && protocol != v1.ProtocolHTTP {
			buf.WriteString(",expose")
		} else if !*e.Expose && protocol == v1.ProtocolHTTP {
			buf.WriteString(",internal")
		}
	}

	if e.HostPort {
		buf.WriteString(",hostport")
	}

	if e.Name != "" {
		if keywords[e.Name] {
			buf.WriteString(",name=")
			buf.WriteString(e.Name)
		} else {
			buf.WriteString(",")
			buf.WriteString(e.Name)
		}
	}

	return buf.String()
}

func ParsePorts(specs ...string) (result []v1.ContainerPort, err error) {
	for _, spec := range specs {
		cp, err := ParsePort(spec)
		if err != nil {
			return nil, err
		}
		result = append(result, cp)
	}
	return
}

func ParsePort(spec string) (result v1.ContainerPort, err error) {
	parts := strings.Split(spec, ",")

	for k, v := range kv.SplitMapFromSlice(parts[1:]) {
		switch k {
		case "name":
			result.Name = v
		case "expose":
			result.Expose = &[]bool{v != "false"}[0]
		case "internal":
			result.Expose = &[]bool{false}[0]
		case "hostport":
			result.HostPort = true
		default:
			result.Name = k
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

	result.Protocol = normalizeProtocol(proto)
	if proto != "" && result.Protocol == "" {
		return result, fmt.Errorf("invalid protocol %s", proto)
	}

	if privStr != "" {
		n, err := strconv.Atoi(privStr)
		if err != nil {
			return result, errors.Wrapf(err, "failed to parse port %s", privStr)
		}
		result.TargetPort = int32(n)
	}

	return
}

func NormalizeContainerPort(port v1.ContainerPort) v1.ContainerPort {
	if port.Port == 0 {
		port.Port = port.TargetPort
	}

	if port.TargetPort == 0 {
		port.TargetPort = port.Port
	}

	if port.Protocol == "" {
		port.Protocol = v1.ProtocolHTTP
	}

	return port
}
