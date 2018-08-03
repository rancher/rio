package deploy

import (
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func destinations(objects []runtime.Object, stack *StackResources) []runtime.Object {
	for _, service := range stack.Services {
		objects = destinationForService(objects, service)
	}

	return objects
}

func destinationForService(objects []runtime.Object, service *v1beta1.Service) []runtime.Object {
	dr := v1alpha3.DestinationRule{
		Host: service.Name,
	}

	dr.Subsets = append(dr.Subsets, &v1alpha3.Subset{
		Name: "latest",
		Labels: map[string]string{
			"rio.cattle.io/revision": "latest",
		},
	})

	for rev := range service.Spec.Revisions {
		dr.Subsets = append(dr.Subsets, &v1alpha3.Subset{
			Name: rev,
			Labels: map[string]string{
				"rio.cattle.io/revision": rev,
			},
		})
	}

	drObject := newDestinationRule(service.Name, service.Namespace)
	drObject.Spec = &dr

	objects = append(objects, drObject)
	return objects
}

func newDestinationRule(name, namespace string) *IstioObject {
	labels := map[string]string{
		"rio.cattle.io":           "true",
		"rio.cattle.io/service":   name,
		"rio.cattle.io/namespace": namespace,
	}

	return &IstioObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}
