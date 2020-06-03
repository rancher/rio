package v1

import (
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/genericcondition"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	CertificateConditionReady = condition.Cond("Ready")
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Certificate is an admin group which manages letsencrypt certificates that are used by ClusterDomain and PublicDomain
type Certificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateSpec   `json:"spec,omitempty"`
	Status CertificateStatus `json:"status,omitempty"`
}

type CertificateSpec struct {
	// SecretRef holds secret reference that stores tls.key and tls.crt
	SecretRef v1.SecretReference `json:"secretRef,omitempty"`

	// DNSNames store SANs used by certificate
	DNSNames []string `json:"dnsNames,omitempty"`
}

type CertificateStatus struct {
	Conditions []genericcondition.GenericCondition `json:"conditions,omitempty"`
}
