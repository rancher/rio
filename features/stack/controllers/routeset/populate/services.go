package populate

import (
	"github.com/rancher/norman/pkg/objectset"
	"github.com/rancher/rio/pkg/namespace"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	v1client "github.com/rancher/types/apis/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func ServiceForRouteSet(r *riov1.RouteSet, stack *riov1.Stack, os *objectset.ObjectSet) error {
	ns, name := namespace.NameRefWithNamespace(r.Name, stack)
	service := v1client.NewService(ns, name,
		v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":                   r.Name,
					"rio.cattle.io/version": "v0",
					"rio.cattle.io/service": r.Name,
					"rio.cattle.io/stack":   stack.Name,
					"rio.cattle.io/project": stack.Namespace,
				},
			},
			Spec: v1.ServiceSpec{
				Type: v1.ServiceTypeClusterIP,
				Ports: []v1.ServicePort{
					{
						Name:       "http-80-80",
						Protocol:   v1.ProtocolTCP,
						Port:       80,
						TargetPort: intstr.FromInt(80),
					},
				},
			},
		})
	os.Add(service)
	return nil
}
