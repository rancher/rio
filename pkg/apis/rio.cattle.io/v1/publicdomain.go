package v1

import (
	genericcondition "github.com/rancher/wrangler/pkg/genericcondition"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PublicDomain is a top-level resource to allow user to its own public domain for the services inside cluster. It can be pointed to
// Router or Service. It is user's responsibility to setup a CNAME or A record to the clusterDomain or ingress IP.
type PublicDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PublicDomainSpec   `json:"spec,inline"`
	Status PublicDomainStatus `json:"status,inline"`
}

type PublicDomainSpec struct {
	// SecretRef reference the secret that contains key and certs for TLS configuration. By default it is configured to use Letsencrypt
	SecretRef v1.SecretReference `json:"secretRef,omitempty"`

	// Whether to disable Letsencrypt certificates.
	DisableLetsencrypt bool `json:"disableLetsencrypt,omitempty"`

	// Target Service Name in the same Namespace
	TargetServiceName string `json:"targetServiceName,omitempty"`

	// PublicDomain name
	DomainName string `json:"domainName,omitempty"`
}

type PublicDomainStatus struct {
	// Whether HTTP is supported in the Domain
	HttpsSupported bool `json:"httpsSupported,omitempty"`

	// Endpoint to access this Domain
	Endpoint string `json:"endpoint,omitempty"`

	// Represents the latest available observations of a PublicDomain's current state.
	Conditions []genericcondition.GenericCondition `json:"conditions,omitempty"`
}
