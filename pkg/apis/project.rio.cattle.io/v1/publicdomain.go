package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PublicDomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PublicDomainSpec `json:"spec,inline"`
}

type PublicDomainSpec struct {
	TargetName      string `json:"targetName,omitempty"`
	TargetStackName string `json:"targetStackName,omitempty"`
	DomainName      string `json:"domainName,omitempty"`
}
