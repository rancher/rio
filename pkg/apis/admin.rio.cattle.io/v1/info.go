package v1

import (
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	RioInfoReady = condition.Cond("Ready")
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RioInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status RioInfoStatus `json:"status,omitempty"`
}

type RioInfoStatus struct {
	Version                 string                              `json:"version,omitempty"`
	GitCommit               string                              `json:"gitCommit,omitempty"`
	SystemNamespace         string                              `json:"systemNamespace,omitempty"`
	Conditions              []genericcondition.GenericCondition `json:"conditions,omitempty"`
	Ready                   bool                                `json:"ready,omitempty"`
	SystemComponentReadyMap map[string]string                   `json:"systemComponentReadyMap,omitempty"`
}
