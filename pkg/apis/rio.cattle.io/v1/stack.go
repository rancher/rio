package v1

import (
	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Stack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StackSpec   `json:"spec,omitempty"`
	Status StackStatus `json:"status,omitempty"`
}

type StackSpec struct {
	Template string            `json:"template,omitempty"`
	Images   map[string]string `json:"images,omitempty"`
	Answers  map[string]string `json:"answers,omitempty"`
}

type StackStatus struct {
	Conditions []genericcondition.GenericCondition `json:"conditions,omitempty"`
}
