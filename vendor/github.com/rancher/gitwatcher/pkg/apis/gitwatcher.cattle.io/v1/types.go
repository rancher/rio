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

type GitWatcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitWatcherSpec   `json:"spec"`
	Status GitWatcherStatus `json:"status"`
}

type GitWatcherSpec struct {
	ReceiverURL                    string            `json:"receiverURL,omitempty"`
	RepositoryURL                  string            `json:"repositoryUrl,omitempty"`
	RepositoryCredentialSecretName string            `json:"repositoryCredentialSecretName,omitempty"`
	GithubWebhookToken             string            `json:"githubWebhookToken,omitempty"`
	Provider                       string            `json:"provider,omitempty"`
	Push                           bool              `json:"push,omitempty"`
	PR                             bool              `json:"pr,omitempty"`
	Branch                         string            `json:"branch,omitempty"`
	Tag                            bool              `json:"tag,omitempty"`
	TagIncludeRegexp               string            `json:"tagInclude,omitempty"`
	TagExcludeRegexp               string            `json:"tagExclude,omitempty"`
	ExecutionLabels                map[string]string `json:"executionLabels,omitempty"`
	Enabled                        bool              `json:"enabled,omitempty"`
	GithubDeployment               bool              `json:"githubDeployment,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type GitCommit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitCommitSpec   `json:"spec,omitempty"`
	Status GitCommitStatus `json:"status,omitempty"`
}

type GitCommitSpec struct {
	Action         string `json:"action,omitempty"`
	Payload        string `json:"payload,omitempty"`
	GitWatcherName string `json:"gitWatcherName,omitempty"`
	Commit         string `json:"commit,omitempty"`
	Branch         string `json:"branch,omitempty"`
	Tag            string `json:"tag,omitempty"`
	PR             string `json:"pr,omitempty"`
	Merged         bool   `json:"merged,omitempty"`
	Closed         bool   `json:"closed,omitempty"`
	SourceLink     string `json:"sourceLink,omitempty"`
	RepositoryURL  string `json:"repositoryUrl,omitempty"`
	Title          string `json:"title,omitempty"`
	Message        string `json:"message,omitempty"`
	Author         string `json:"author,omitempty"`
	AuthorEmail    string `json:"authorEmail,omitempty"`
	AuthorAvatar   string `json:"authorAvatar,omitempty"`
}

type GitWatcherStatus struct {
	Conditions  []Condition `json:"conditions,omitempty"`
	Token       string      `json:"token,omitempty"`
	HookID      string      `json:"hookId,omitempty"`
	FirstCommit string      `json:"firstCommit,omitempty"`
}

type GithubStatus struct {
	DeploymentID    int64  `json:"deploymentId,omitempty"`
	DeploymentState string `json:"deploymentState,omitempty"`
	EnvironmentURL  string `json:"environmentUrl,omitempty"`
	LogURL          string `json:"logUrl,omitempty"`
}

type GitCommitStatus struct {
	Conditions    []Condition   `json:"conditions,omitempty"`
	StatusURL     string        `json:"statusUrl,omitempty"`
	AppliedStatus string        `json:"appliedStatus,omitempty"`
	BuildStatus   string        `json:"buildStatus,omitempty"`
	GithubStatus  *GithubStatus `json:"githubStatus,omitempty"`
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
