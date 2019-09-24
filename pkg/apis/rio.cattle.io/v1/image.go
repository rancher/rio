package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ImageBuild is used to build an image from a Git source and optionally push it
type ImageBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageBuildSpec   `json:"spec,omitempty"`
	Status ImageBuildStatus `json:"status,omitempty"`
}

type ImageBuildSpec struct {
	// Repository url
	Repo string `json:"repo,omitempty"`

	// Repo Revision. Can be a git commit or tag
	Revision string `json:"revision,omitempty"`

	// Repo Branch. If specified, a gitmodule will be created to watch the repo and creating new revision if new commit or tag is pushed.
	Branch string `json:"branch,omitempty"`

	// Specify the name of the Dockerfile in the Repo. This is the full path relative to the repo root. Defaults to `Dockerfile`.
	Dockerfile string `json:"dockerfile,omitempty"`

	// Specify build context. Defaults to "."
	Context string `json:"context,omitempty"`

	// Specify build args
	Args []string `json:"args,omitempty"`

	// Specify the build template. Defaults to `buildkit`.
	Template string `json:"template,omitempty"`

	// Specify the github secret name. Used to create Github webhook, the secret key has to be `accessToken`
	WebhookSecretName string `json:"webhookSecretName,omitempty"`

	// Specify secret name for checking our git resources
	PullSecretName string `json:"pullSecretName,omitempty"`

	// Specify custom registry to push the image instead of built-in one
	PushRegistry string `json:"pushRegistry,omitempty"`

	// Specify secret for pushing to custom registry
	PushRegistrySecretName string `json:"pushRegistrySecretName,omitempty"`

	// Specify image name instead of the one generated from service name, format: $registry/$imageName:$revision
	ImageName string `json:"imageName,omitempty"`

	// Whether to enable builds for pull requests
	PR bool `json:"pr,omitempty"`

	// Whether to enable builds for tags
	Tag bool `json:"tag,omitempty"`

	// Build image with no cache
	NoCache bool `json:"noCache,omitempty"`

	// Timeout describes how long the build can run
	Timeout *metav1.Duration
}

type ImageBuildStatus struct {
	ImageName string
}
