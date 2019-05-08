package v1

import (
	"github.com/rancher/rio/pkg/apis/common"
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
	Public          bool         `json:"public,omitempty"`
	ServiceName     string       `json:"serviceName,omitempty"`
	Version         string       `json:"Version,omitempty"`
	AdjustedWeight  int          `json:"adjustedWeight,omitempty"`
	Weight          int          `json:"weight,omitempty"`
	Scale           int          `json:"scale,omitempty"`
	ScaleStatus     *ScaleStatus `json:"scaleStatus,omitempty"`
	DeploymentReady bool         `json:"deploymentReady,omitempty"`
	RolloutConfig
}

type ServiceObservedWeight struct {
	LastWrite   metav1.Time `json:"lastWrite,omitempty"`
	Weight      int         `json:"weight,omitempty"`
	ServiceName string      `json:"serviceName,omitempty"`
}

type AppStatus struct {
	PublicDomains  []string                            `json:"publicDomains,omitempty"`
	Endpoints      []string                            `json:"endpoints,omitempty"`
	Conditions     []genericcondition.GenericCondition `json:"conditions,omitempty"`
	RevisionWeight map[string]ServiceObservedWeight    `json:"revisionWeight,omitempty"`
}

func (in *App) State() common.State {
	return common.StateFromConditionAndMeta(in.ObjectMeta, in.Status.Conditions)
}
