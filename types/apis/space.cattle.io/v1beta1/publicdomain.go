package v1beta1

import (
	"github.com/rancher/norman/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PublicDomain struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PublicDomainSpec `json:"spec,inline"`
}

type PublicDomainSpec struct {
	TargetName          string `json:"targetName,omitempty"`
	TargetStackName     string `json:"targetStackName,omitempty"`
	TargetWorkspaceName string `json:"targetWorkspaceName,omitempty"`
	DomainName          string `json:"domainName,omitempty"`
}
