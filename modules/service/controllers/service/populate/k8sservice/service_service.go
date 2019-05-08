package k8sservice

import (
	"github.com/rancher/rio/modules/service/controllers/service/populate/servicelabels"
	"github.com/rancher/rio/modules/service/controllers/service/populate/serviceports"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func serviceSelector(service *riov1.Service, os *objectset.ObjectSet) {
	labels := servicelabels.ServiceLabels(service)
	selectorLabels := servicelabels.SelectorLabels(service)
	svc := newServiceSelector(service.Name, service.Namespace, labels, selectorLabels)
	if len(serviceports.ServiceNamedPorts(service)) > 0 {
		svc.Spec.Ports = serviceports.ServiceNamedPorts(service)
	}
	os.Add(svc)
}

func newServiceSelector(name, namespace string, labels, selectorLabels map[string]string) *v1.Service {
	return constructors.NewService(namespace, name, v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
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
	})
}
