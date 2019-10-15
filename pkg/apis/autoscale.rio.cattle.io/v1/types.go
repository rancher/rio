package v1

import (
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ServiceScaleRecommendationSynced = condition.Cond("Synced")
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ServiceScaleRecommendation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceScaleRecommendationSpec   `json:"spec,omitempty"`
	Status ServiceScaleRecommendationStatus `json:"status,omitempty"`
}

type ServiceScaleRecommendationSpec struct {
	MinScale    int                   `json:"minScale,omitempty"`
	MaxScale    int                   `json:"maxScale,omitempty"`
	Concurrency int                   `json:"concurrency,omitempty"`
	Replicas    *int32                `json:"replicas,omitempty"`
	Selector    *metav1.LabelSelector `json:"selector"`
}

type ServiceScaleRecommendationStatus struct {
	Conditions []genericcondition.GenericCondition `json:"conditions,omitempty"`
}
