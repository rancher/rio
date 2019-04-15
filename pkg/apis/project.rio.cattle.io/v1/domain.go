package v1

import (
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/genericcondition"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ClusterDomainConditionReady = condition.Cond("Ready")
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterDomainSpec   `json:"spec,omitempty"`
	Status ClusterDomainStatus `json:"status,omitempty"`
}

type ClusterDomainSpec struct {
	SecretRef  v1.SecretReference
	Addresses  []Address   `json:"addresses,omitempty"`
	Subdomains []Subdomain `json:"subdomains,omitempty"`
}

type Address struct {
	IP string `json:"ip,omitempty"`
}

type Subdomain struct {
	Name      string    `json:"name,omitempty"`
	Addresses []Address `json:"addresses,omitempty"`
}

type ClusterDomainStatus struct {
	HTTPSSupported bool                                `json:"httpsSupported,omitempty"`
	ClusterDomain  string                              `json:"domain,omitempty"`
	Conditions     []genericcondition.GenericCondition `json:"conditions,omitempty"`
}
