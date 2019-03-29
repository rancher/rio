package v1

import (
	"github.com/rancher/wrangler/pkg/condition"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GitWebHookReceiverConditionRegistered   condition.Cond = "Registered"
	GitWebHookExecutionConditionInitialized condition.Cond = "Initialized"
	GitWebHookExecutionConditionHandled     condition.Cond = "Handled"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GitWebHookReceiver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitWebHookReceiverSpec   `json:"spec"`
	Status GitWebHookReceiverStatus `json:"status"`
}

type GitWebHookReceiverSpec struct {
	RepositoryURL                  string            `json:"repositoryUrl,omitempty"`
	RepositoryCredentialSecretName string            `json:"repositoryCredentialSecretName,omitempty"`
	Provider                       string            `json:"provider,omitempty"`
	Push                           bool              `json:"push,omitempty"`
	PR                             bool              `json:"pr,omitempty"`
	Tag                            bool              `json:"tag,omitempty"`
	ExecutionLabels                map[string]string `json:"executionLabels,omitempty"`
	Enabled                        bool              `json:"enabled,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GitWebHookExecution struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitWebHookExecutionSpec   `json:"spec,omitempty"`
	Status GitWebHookExecutionStatus `json:"status,omitempty"`
}

type GitWebHookExecutionSpec struct {
	Payload                string `json:"payload,omitempty"`
	GitWebHookReceiverName string `json:"gitWebHookReceiverName,omitempty"`
	Commit                 string `json:"commit,omitempty"`
	Branch                 string `json:"branch,omitempty"`
	Tag                    string `json:"tag,omitempty"`
	PR                     string `json:"pr,omitempty"`
	SourceLink             string `json:"sourceLink,omitempty"`
	RepositoryURL          string `json:"repositoryUrl,omitempty"`
	Title                  string `json:"title,omitempty"`
	Message                string `json:"message,omitempty"`
	Author                 string `json:"author,omitempty"`
	AuthorEmail            string `json:"authorEmail,omitempty"`
	AuthorAvatar           string `json:"authorAvatar,omitempty"`
}

type GitWebHookReceiverStatus struct {
	Conditions []Condition `json:"conditions,omitempty"`
	Token      string      `json:"token,omitempty"`
	HookID     string      `json:"hookId,omitempty"`
}

type GitWebHookExecutionStatus struct {
	Conditions    []Condition `json:"conditions,omitempty"`
	StatusURL     string      `json:"statusUrl,omitempty"`
	AppliedStatus string      `json:"appliedStatus,omitempty"`
}

type Condition struct {
	// Type of the condition.
	Type string `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition
	Message string `json:"message,omitempty"`
}
