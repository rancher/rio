package v1

import (
	"github.com/rancher/types/apis/management.cattle.io/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Feature struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec FeatureSpec `json:"spec,omitempty"`
}

type FeatureSpec struct {
	Description string            `json:"description,omitempty"`
	Enable      bool              `json:"enable,omitempty"`
	Questions   []v3.Question     `json:"questions,omitempty"`
	Answers     map[string]string `json:"answers,omitempty"`
}
