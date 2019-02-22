package create

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/norman/pkg/kv"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func ParseExposedPorts(specs []string) ([]riov1.ExposedPort, error) {
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

func ParsePorts(specs []string) ([]riov1.PortBinding, error) {
	var result []riov1.PortBinding

	for _, spec := range specs {
		portBinding, err := parsePortBinding(spec)
		if err != nil {
			return nil, err
		}
		result = append(result, portBinding)
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
