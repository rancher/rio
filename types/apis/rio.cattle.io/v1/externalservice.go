package v1

import (
	"github.com/rancher/norman/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExternalService struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ExternalServiceSpec `json:"spec,omitempty"`
}

type ExternalServiceSpec struct {
	Target string `json:"target,omitempty"`
	StackScoped
}
