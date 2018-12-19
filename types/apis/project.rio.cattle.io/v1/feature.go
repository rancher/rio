package v1

import (
	"github.com/rancher/norman/condition"
	"github.com/rancher/norman/types"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	FeatureConditionEnabled = condition.Cond("Enabled")
)

type Feature struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FeatureSpec   `json:"spec,omitempty"`
	Status FeatureStatus `json:"status,omitempty"`
}

type FeatureSpec struct {
	Description string            `json:"description,omitempty"`
	Enabled     bool              `json:"enable,omitempty"`
	Questions   []v3.Question     `json:"questions,omitempty"`
	Answers     map[string]string `json:"answers,omitempty"`
}

type FeatureStatus struct {
	Conditions []condition.GenericCondition `json:"conditions,omitempty"`
}
