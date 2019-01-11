package v1alpha3

import (
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/rancher/norman/types"
	istiov1alpha3 "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Gateway struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec v1alpha3.GatewaySpec `json:"spec"`
}

type VirtualService struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec v1alpha3.VirtualServiceSpec `json:"spec"`
}

type ServiceEntry struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ServiceEntrySpec `json:"spec"`
}

type DestinationRule struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec v1alpha3.DestinationRuleSpec `json:"spec"`
}

// Copied from istio

type ServiceEntrySpec struct {
	Hosts      []string                              `json:"hosts,omitempty"`
	Addresses  []string                              `json:"addresses,omitempty"`
	Ports      []Port                                `json:"ports,omitempty"`
	Location   istiov1alpha3.ServiceEntry_Location   `json:"location,omitempty"`
	Resolution istiov1alpha3.ServiceEntry_Resolution `json:"resolution,omitempty"`
	Endpoints  []ServiceEntry_Endpoint               `json:"endpoints,omitempty"`
}

type ServiceEntry_Endpoint struct {
	Address string            `json:"address,omitempty"`
	Ports   map[string]uint32 `json:"ports,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}

type Port struct {
	Number   uint32 `json:"number,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Name     string `json:"name,omitempty"`
}
