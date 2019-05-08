package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GitModule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitModuleSpec   `json:"spec,omitempty"`
	Status GitModuleStatus `json:"status,omitempty"`
}

type GitModuleSpec struct {
	ServiceName      string `json:"serviceName,omitempty"`
	ServiceNamespace string `json:"serviceNamespace,omitempty"`
	Repo             string `json:"repo,omitempty"`
	Secret           string `json:"secret,omitempty"`
	Branch           string `json:"branch,omitempty"`
}

type GitModuleStatus struct {
	LastRevision string `json:"lastRevision,omitempty"`
}
