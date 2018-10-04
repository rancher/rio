package v1beta1

import (
	"github.com/rancher/norman/condition"
	"github.com/rancher/norman/types"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	StackConditionDefined  = condition.Cond("Defined")
	StackConditionDeployed = condition.Cond("Deployed")
)

type Stack struct {
	types.Namespaced

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StackSpec   `json:"spec"`
	Status StackStatus `json:"status"`
}

type StackSpec struct {
	Description               string            `json:"description,omitempty"`
	Template                  string            `json:"template,omitempty"`
	AdditionalFiles           map[string]string `json:"additionalFiles,omitempty"`
	Answers                   map[string]string `json:"answers,omitempty"`
	Questions                 []v3.Question     `json:"questions,omitempty"`
	DisableMesh               bool              `json:"disableMesh,omitempty"`
	EnableKubernetesResources bool              `json:"enableKubernetesResources,omitempty"`
	RepoTag                   []string          `json:"repoTag,omitempty"`
}

type StackStatus struct {
	Conditions []Condition `json:"conditions,omitempty"`
}

type StackScoped struct {
	StackName string `json:"stackName,omitempty" norman:"type=reference[stack],required,noupdate"`
	SpaceName string `json:"spaceName,omitempty" norman:"type=reference[/v1beta1-rio/schemas/space],noupdate"`
}

type InternalStack struct {
	Services   map[string]Service  `json:"services,omitempty"`
	Configs    map[string]Config   `json:"configs,omitempty"`
	Volumes    map[string]Volume   `json:"volumes,omitempty"`
	Routes     map[string]RouteSet `json:"routes,omitempty"`
	Kubernetes Kubernetes          `json:"kubernetes,omitempty"`
}
