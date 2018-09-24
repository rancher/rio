package k8sservice

import (
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/servicelabels"
	"github.com/rancher/rio/pkg/deploy/stack/populate/serviceports"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func serviceSelector(stack *input.Stack, service *v1beta1.Service, output *output.Deployment) {
	labels := servicelabels.ServiceLabels(stack, service)
	selectorLabels := servicelabels.SelectorLabels(stack, service)
	svc := newServiceSelector(service.Name, stack.Namespace, labels, selectorLabels)
	ports, ip := serviceports.ServiceNamedPorts(service)

	if len(ports) > 0 {
		svc.Spec.Ports = ports
	}

	if service.Spec.Revision.ParentService == "" {
		nonVersioned := svc.DeepCopy()
		nonVersioned.Name = service.Spec.Revision.ServiceName
		output.Services[nonVersioned.Name] = nonVersioned

		if ip != "" {
			svc.Spec.ClusterIP = ip
		}
		delete(svc.Spec.Selector, "rio.cattle.io/version")
	}

	output.Services[svc.Name] = svc
}

func newServiceSelector(name, namespace string, labels, selectorLabels map[string]string) *v1.Service {
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
			Selector: selectorLabels,
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
