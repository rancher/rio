package v1

import (
	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Stack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StackSpec   `json:"spec,omitempty"`
	Status StackStatus `json:"status,omitempty"`
}

type StackSpec struct {
	// Stack build parameters that watches git repo
	Build *StackBuild `json:"build,omitempty"`

	// Stack template
	Template string `json:"template,omitempty"`

	// Stack images
	Images map[string]string `json:"images,omitempty"`

	// Stack answers
	Answers map[string]string `json:"answers,omitempty"`
}

type StackBuild struct {
	// Git repo url
	Repo string `json:"repo,omitempty"`

	// Git branch
	Branch string `json:"branch,omitempty"`

	// Git revision
	Revision string `json:"revision,omitempty"`

	// Git secret name for repository
	CloneSecretName string `json:"cloneSecretName,omitempty"`
}

type StackStatus struct {
	// Observed commit for the build
	Revision string `json:"revision,omitempty"`

	Conditions []genericcondition.GenericCondition `json:"conditions,omitempty"`
}
