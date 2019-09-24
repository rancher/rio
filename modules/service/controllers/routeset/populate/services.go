package populate

import (
	"fmt"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func ServiceForRouteSet(r *riov1.Router, os *objectset.ObjectSet) error {
	service := constructors.NewService(r.Namespace, r.Name,
		v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app":     r.Name,
					"version": "v0",
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

	if constants.ServiceMeshMode == constants.ServiceMeshModeLinkerd {
		for i := range r.Spec.Routes {
			os.Add(constructors.NewService(r.Namespace, fmt.Sprintf("%s-%v", r.Name, i),
				v1.Service{
					ObjectMeta: metav1.ObjectMeta{},
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
				}))
		}
	}
	os.Add(service)
	return nil
}
