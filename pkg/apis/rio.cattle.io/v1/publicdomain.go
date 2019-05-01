package v1

import (
	genericcondition "github.com/rancher/wrangler/pkg/genericcondition"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PublicDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PublicDomainSpec   `json:"spec,inline"`
	Status PublicDomainStatus `json:"status,inline"`
}

type PublicDomainSpec struct {
	SecretRef          v1.SecretReference `json:"secretRef,omitempty"`
	DisableLetsencrypt bool               `json:"disableLetsencrypt,omitempty"`
	TargetServiceName  string             `json:"targetServiceName,omitempty"`
	DomainName         string             `json:"domainName,omitempty"`
}

type PublicDomainStatus struct {
	HttpsSupported bool                                `json:"httpsSupported,omitempty"`
	Endpoint       string                              `json:"endpoint,omitempty"`
	Conditions     []genericcondition.GenericCondition `json:"conditions,omitempty"`
}
