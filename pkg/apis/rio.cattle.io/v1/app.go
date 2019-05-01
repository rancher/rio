package v1

import (
	genericcondition "github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSpec   `json:"spec,omitempty"`
	Status AppStatus `json:"status,omitempty"`
}

type AppSpec struct {
	Revisions []Revision `json:"revisions,omitempty"`
}

type Revision struct {
	ServiceName     string
	Version         string
	AdjustedWeight  int
	Weight          int
	DeploymentReady bool
	RolloutConfig
}

type ServiceObservedWeight struct {
	LastWrite   metav1.Time
	Weight      int
	ServiceName string
}

type AppStatus struct {
	Conditions     []genericcondition.GenericCondition `json:"conditions,omitempty"`
	RevisionWeight map[string]ServiceObservedWeight    `json:"weight,omitempty"`
}
