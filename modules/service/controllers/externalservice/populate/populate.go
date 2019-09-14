package populate

import (
	"fmt"
	"strconv"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/parse"
	"github.com/rancher/wrangler/pkg/objectset"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ServiceForExternalService(es *riov1.ExternalService, namespace *corev1.Namespace, os *objectset.ObjectSet) error {
	svc := constructors.NewService(namespace.Name, es.Name, v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"rio.cattle.io/service": es.Name,
			},
		},
	})
	if es.Spec.FQDN != "" {
		u, err := parse.TargetURL(es.Spec.FQDN)
		if err != nil {
			return err
		}
		svc.Spec = v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: u.Hostname(),
		}
	} else if len(es.Spec.IPAddresses) > 0 {
		var hosts []string
		var ports []int32
		svc.Spec = v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: v1.ClusterIPNone,
		}
		for _, ip := range es.Spec.IPAddresses {
			u, err := parse.TargetURL(ip)
			if err != nil {
				return err
			}
			port := u.Port()
			if port == "" {
				port = "80"
			}
			portInt, _ := strconv.Atoi(port)
			svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
				Name:     fmt.Sprintf("http-%v-%v", portInt, portInt),
				Protocol: v1.ProtocolTCP,
				Port:     int32(portInt),
			})
			hosts = append(hosts, u.Hostname())
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
					Name:     fmt.Sprintf("http-%v-%v", ports[i], ports[i]),
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
