package k8sservice

import (
	"github.com/rancher/rio/modules/service/controllers/service/populate/servicelabels"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/serviceports"
	"github.com/rancher/rio/pkg/services"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func serviceSelector(service *riov1.Service, os *objectset.ObjectSet) {
	labels := servicelabels.ServiceLabels(service)
	selectorLabels := servicelabels.SelectorLabels(service)
	app, version := services.AppAndVersion(service)
	svc := newServiceSelector(app+"-"+version, service.Namespace, v1.ServiceTypeClusterIP, labels, selectorLabels)
	if len(serviceports.ServiceNamedPorts(service)) > 0 {
		svc.Spec.Ports = serviceports.ServiceNamedPorts(service)
	}
	os.Add(svc)
}

func newServiceSelector(name, namespace string, serviceType v1.ServiceType, labels, selectorLabels map[string]string) *v1.Service {
	return constructors.NewService(namespace, name, v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Type:     serviceType,
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
