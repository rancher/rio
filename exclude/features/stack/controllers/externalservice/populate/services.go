package populate

import (
	"fmt"
	"strconv"

	"github.com/rancher/rio/features/routing/controllers/externalservice/populate"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ServiceForExternalService(es *riov1.ExternalService, stack *riov1.Stack, os *objectset.ObjectSet) error {
	svc := constructors.NewService(stack.Name, es.Name, v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"rio.cattle.io/service": es.Name,
				"rio.cattle.io/stack":   stack.Name,
				"rio.cattle.io/project": stack.Namespace,
			},
		},
	})
	if es.Spec.FQDN != "" {
		u, err := populate.ParseTargetUrl(es.Spec.FQDN)
		if err != nil {
			return err
		}
		svc.Spec = v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: u.Host,
		}
	} else if len(es.Spec.IPAddresses) > 0 {
		var hosts []string
		var ports []int32
		for _, ip := range es.Spec.IPAddresses {
			u, err := populate.ParseTargetUrl(ip)
			if err != nil {
				return err
			}
			port := u.Port()
			if port == "" {
				port = "80"
			}
			portInt, _ := strconv.Atoi(port)
			svc.Spec = v1.ServiceSpec{
				Type: v1.ServiceTypeClusterIP,
				Ports: []v1.ServicePort{
					{
						Name:     fmt.Sprintf("http-%v-%v", portInt, portInt),
						Protocol: v1.ProtocolTCP,
						Port:     int32(portInt),
					},
				},
			}
			hosts = append(hosts, u.Host)
			ports = append(ports, int32(portInt))
		}
		os.Add(populateEndpoint(svc.Name, svc.Namespace, hosts, ports))
	} else if es.Spec.Service != "" {
		svc.Spec = v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name:     "http-80-80",
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
			},
		}
	}

	os.Add(svc)
	return nil
}

func populateEndpoint(name, namespace string, hosts []string, ports []int32) *v1.Endpoints {
	var subnet []v1.EndpointSubset
	for i, host := range hosts {
		subnet = append(subnet, v1.EndpointSubset{
			Addresses: []v1.EndpointAddress{
				{
					IP: host,
				},
			},
			Ports: []v1.EndpointPort{
				{
					Name:     "http-80-80",
					Protocol: v1.ProtocolTCP,
					Port:     ports[i],
				},
			},
		})
	}
	return constructors.NewEndpoints(namespace, name, v1.Endpoints{
		Subsets: subnet,
	})
}
