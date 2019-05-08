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
	ServiceNameToRead string            `json:"serviceNameToRead,omitempty"`
	ZeroScaleService  string            `json:"zeroScaleService,omitempty"`
	MinScale          int32             `json:"minScale,omitempty"`
	MaxScale          int32             `json:"maxScale,omitempty"`
	Concurrency       int               `json:"concurrency,omitempty"`
	PrometheusURL     string            `json:"prometheusURL,omitempty"`
	Selector          map[string]string `json:"selector,omitempty"`
}

type ServiceScaleRecommendationStatus struct {
	DesiredScale *int32                              `json:"desiredScale,omitempty"`
	Conditions   []genericcondition.GenericCondition `json:"conditions,omitempty"`
}
