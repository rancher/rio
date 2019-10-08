package v1

import (
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ClusterDomainConditionReady = condition.Cond("Ready")
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ClusterDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterDomainSpec   `json:"spec,omitempty"`
	Status ClusterDomainStatus `json:"status,omitempty"`
}

type ClusterDomainSpec struct {
	// SecretName holding the TLS certificate for this domain.  This is expected
	// to be a wildcard certificate
	SecretName string `json:"secretName,omitempty"`

	// The public HTTPS port for the cluster domain
	HTTPSPort int `json:"httpsPort,omitempty"`
	// The public HTTP port for the cluster domain
	HTTPPort int `json:"httpPort,omitempty"`

	// The addresses assigned to the ClusterDomain by the provider
	Addresses []Address `json:"addresses,omitempty"`
}

type Address struct {
	IP       string `json:"ip,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

type ClusterDomainStatus struct {
	AssignedSecretName string                              `json:"assignedSecretName,omitempty"`
	HTTPSSupported     bool                                `json:"httpsSupported,omitempty"`
	Conditions         []genericcondition.GenericCondition `json:"conditions,omitempty"`
}
