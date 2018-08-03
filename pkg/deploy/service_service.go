package deploy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func serviceNamedPorts(service *v1beta1.ServiceUnversionedSpec) ([]v1.ServicePort, string) {
	var result []v1.ServicePort

	eps := service.ExposedPorts
	for _, k := range sortKeys(service.Sidekicks) {
		container := service.Sidekicks[k]
		eps = append(eps, container.ExposedPorts...)
	}

	for _, portBindings := range service.PortBindings {
		eps = append(eps, v1beta1.ExposedPort{
			PortBinding: v1beta1.PortBinding{
				TargetPort: portBindings.TargetPort,
				Port:       portBindings.TargetPort,
				Protocol:   portBindings.Protocol,
			},
		})
	}

	ip := ""
	portsDefined := map[string]bool{}
	names := map[string]bool{}
	for _, port := range eps {
		if port.Port == 0 {
			port.Port = port.TargetPort
		}

		if port.IP != "" {
			ip = port.IP
		}

		name := ""
		defName := fmt.Sprintf("%s-%d-%d", port.Protocol, port.Port, port.TargetPort)
		if port.Name == "" {
			name = defName
		} else {
			name = port.Name
		}
		if names[name] || portsDefined[defName] {
			continue
		}

		portsDefined[defName] = true
		names[name] = true

		servicePort := v1.ServicePort{
			Name:       name,
			TargetPort: intstr.FromInt(int(port.TargetPort)),
			Port:       int32(port.Port),
			Protocol:   v1.ProtocolTCP,
		}

		if strings.EqualFold(port.Protocol, "udp") {
			servicePort.Protocol = v1.ProtocolUDP
		}

		result = append(result, servicePort)
	}

	return result, ip
}

func serviceSelector(objects []runtime.Object, name, namespace string, service *v1beta1.ServiceUnversionedSpec, labels map[string]string) []runtime.Object {
	svc := newServiceSelector(name, namespace, labels)
	ports, ip := serviceNamedPorts(service)

	if len(ports) > 0 {
		svc.Spec.Ports = ports
	}

	if ip != "" {
		svc.Spec.ClusterIP = ip
	}

	objects = append(objects, svc)
	return objects
}

func newServiceSelector(name, namespace string, labels map[string]string) *v1.Service {
	// for "latest" selector we want all revisions
	if labels["rio.cattle.io/revision"] == "latest" {
		newLabels := map[string]string{}
		for k, v := range labels {
			newLabels[k] = v
		}
		delete(newLabels, "rio.cattle.io/revision")
		labels = newLabels
	}

	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []v1.ServicePort{
				{
					Name:       "default",
					Protocol:   v1.ProtocolTCP,
					TargetPort: intstr.FromInt(80),
					Port:       80,
				},
			},
		},
	}
}

func sortKeys(m map[string]v1beta1.SidekickConfig) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
