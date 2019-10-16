package v1

import (
	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PublicDomain is a top-level resource to allow user to its own public domain for the services inside cluster. It can be pointed to
// Router or Service. It is user's responsibility to setup a CNAME or A record to the clusterDomain or ingress IP.
type PublicDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PublicDomainSpec   `json:"spec,omitempty"`
	Status PublicDomainStatus `json:"status,omitempty"`
}

type PublicDomainSpec struct {
	// SecretName holding the TLS certificate for this domain.
	SecretName string `json:"secretName,omitempty"`

	// Target App Name.  Can be a Router name also
	TargetApp string `json:"targetApp,omitempty"`

	// Target Version
	TargetVersion string `json:"targetVersion,omitempty"`

	// Target Service or Router Namespace
	TargetNamespace string `json:"targetNamespace,omitempty"`
}

type PublicDomainStatus struct {
	// Whether HTTP is supported in the Domain
	HTTPSSupported bool `json:"httpsSupported,omitempty"`

	// Secret containing TLS cert for HTTPS
	AssignedSecretName string `json:"assignedSecretName,omitempty"`

	// Represents the latest available observations of a PublicDomain's current state.
	Conditions []genericcondition.GenericCondition `json:"conditions,omitempty"`
}
