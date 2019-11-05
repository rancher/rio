package v1

import (
	"github.com/rancher/wrangler/pkg/genericcondition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	// Permissions used while deploying objects created by this stack
	Permissions []Permission `json:"permissions,omitempty" mapper:"permissions,alias=permission"`

	// Additional GVKs not in the rio.cattle.io that have the rio.cattle.io/stack label. These objects
	// are "owned" by this stack
	AdditionalGroupVersionKinds []schema.GroupVersionKind `json:"additionalGroupVersionKinds,omitempty"`

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

	// Specify the name of the Riofile in the Repo. This is the full path relative to the repo root. Defaults to `Riofile`.
	Riofile string `json:"rioFile,omitempty"`

	// Specify the github secret name. Used to create Github webhook, the secret key has to be `accessToken`
	WebhookSecretName string `json:"webhookSecretName,omitempty"`
}

type StackStatus struct {
	// Observed commit for the build
	Revision string `json:"revision,omitempty"`

	Conditions []genericcondition.GenericCondition `json:"conditions,omitempty"`
}
