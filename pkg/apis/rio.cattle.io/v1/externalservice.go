package v1

import (
	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ExternalService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ExternalServiceSpec   `json:"spec,omitempty"`
	Status            ExternalServiceStatus `json:"status,omitempty"`
}

type ExternalServiceSpec struct {
	// External service located outside mesh, represented by IPs
	IPAddresses []string `json:"ipAddresses,omitempty"`
	// External service located outside mesh, represented by DNS
	FQDN string `json:"fqdn,omitempty"`
	// In-Mesh service name in another stack in current project
	Service string `json:"service,omitempty"`
}

type ExternalServiceStatus struct {
	Conditions []genericcondition.GenericCondition `json:"conditions,omitempty"`
}
