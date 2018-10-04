package istio

import (
	"github.com/rancher/rio/pkg/deploy/stack/input"
	"github.com/rancher/rio/pkg/deploy/stack/output"
	"github.com/rancher/rio/pkg/deploy/stack/populate/service"
	"github.com/rancher/rio/pkg/deploy/stack/populate/servicelabels"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func destinationRules(stack *input.Stack) ([]*output.IstioObject, error) {
	var result []*output.IstioObject
	services, err := service.CollectionServices(stack.Services)
	if err != nil {
		return nil, err
	}

	for name, service := range services {
		result = append(result, destinationRuleForService(stack, name, service))
	}

	return result, nil
}

func destinationRuleForService(stack *input.Stack, name string, service *output.ServiceSet) *output.IstioObject {
	dr := v1alpha3.DestinationRule{
		Host: name,
	}

	dr.Subsets = append(dr.Subsets, newSubSet(service.Service))

	for _, rev := range service.Revisions {
		dr.Subsets = append(dr.Subsets, newSubSet(rev))
	}

	drObject := newDestinationRule(stack, service.Service)
	drObject.Spec = &dr

	return drObject
}

func newSubSet(service *v1beta1.Service) *v1alpha3.Subset {
	return &v1alpha3.Subset{
		Name: service.Spec.Revision.Version,
		Labels: map[string]string{
			"rio.cattle.io/version": service.Spec.Revision.Version,
		},
	}
}

func newDestinationRule(stack *input.Stack, service *v1beta1.Service) *output.IstioObject {
	labels := servicelabels.RioOnlyServiceLabels(stack, service)
	return &output.IstioObject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.istio.io/v1alpha3",
			Kind:       "DestinationRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: service.Namespace,
			Labels:    labels,
		},
	}
}
