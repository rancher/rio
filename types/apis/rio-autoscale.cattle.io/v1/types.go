package v1

import (
	"github.com/rancher/norman/condition"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ServiceScaleRecommendationSynced = condition.Cond("Synced")
)

type ServiceScaleRecommendation struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceScaleRecommendationSpec   `json:"spec,omitempty"`
	Status ServiceScaleRecommendationStatus `json:"status,omitempty"`
}

type ServiceScaleRecommendationSpec struct {
	ServiceNameToRead string `json:"serviceNameToRead,omitempty"`
	ZeroScaleService  string `json:"zeroScaleService,omitempty"`
	MinScale          int32  `json:"minScale,omitempty"`
	MaxScale          int32  `json:"maxScale,omitempty"`
	Concurrency       int    `json:"concurrency,omitempty"`
	PrometheusURL     string `json:"prometheusURL,omitempty"`
}

type ServiceScaleRecommendationStatus struct {
	DesiredScale *int32         `json:"desiredScale,omitempty"`
	Conditions   []v1.Condition `json:"conditions,omitempty"`
}
