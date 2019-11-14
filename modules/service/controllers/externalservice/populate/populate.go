package populate

import (
	"fmt"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/wrangler/pkg/objectset"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ServiceForExternalService(es *riov1.ExternalService, os *objectset.ObjectSet) error {
	if spec, ok := getSpec(es, os); ok {
		svc := constructors.NewService(es.Namespace, es.Name, v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"rio.cattle.io/service": es.Name,
				},
			},
			Spec: spec,
		})
		os.Add(svc)
	}

	return nil
}

func getSpec(es *riov1.ExternalService, os *objectset.ObjectSet) (v1.ServiceSpec, bool) {
	if es.Spec.FQDN != "" {
		return v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: es.Spec.FQDN,
		}, true
	}

	if len(es.Spec.IPAddresses) > 0 {
		os.Add(populateEndpoint(es.Namespace, es.Name, es.Spec.IPAddresses))
		return v1.ServiceSpec{
			Type:      v1.ServiceTypeClusterIP,
			ClusterIP: v1.ClusterIPNone,
		}, true
	}

	service := es.Spec.TargetRouter
	if es.Spec.TargetApp != "" {
		service = es.Spec.TargetApp
		if es.Spec.TargetVersion != "" {
			service = service + "-" + es.Spec.TargetVersion
		}
	}

	if service != "" {
		ns := es.Spec.TargetServiceNamespace
		if ns == "" {
			ns = es.Namespace
		}
		return v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: fmt.Sprintf("%s.%s.svc.cluster.local", service, ns),
		}, true
	}

	return v1.ServiceSpec{}, false
}

func populateEndpoint(namespace, name string, hosts []string) *v1.Endpoints {
	var subnet []v1.EndpointSubset
	for _, host := range hosts {
		subnet = append(subnet, v1.EndpointSubset{
			Addresses: []v1.EndpointAddress{
				{
					IP: host,
				},
			},
		})
	}
	return constructors.NewEndpoints(namespace, name, v1.Endpoints{
		Subsets: subnet,
	})
}
