package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	Args []string `json:"args,omitempty" mapper:"alias=arg"`

	// Specify the build template. Defaults to `buildkit`.
	Template string `json:"template,omitempty"`

	// Specify the github secret name. Used to create Github webhook, the secret key has to be `accessToken`
	WebhookSecretName string `json:"webhookSecretName,omitempty"`

	// Specify secret name for checking our git resources
	CloneSecretName string `json:"pullSecretName,omitempty"`

	// Specify custom registry to push the image instead of built-in one
	PushRegistry string `json:"pushRegistry,omitempty"`

	// Specify secret for pushing to custom registry
	PushRegistrySecretName string `json:"pushRegistrySecretName,omitempty"`

	// Specify image name instead of the one generated from service name, format: $registry/$imageName:$revision
	ImageName string `json:"imageName,omitempty"`

	// Whether to enable builds for pull requests
	PR bool `json:"pr,omitempty" mapper:"alias=onPR"`

	// Whether to enable builds for tags
	Tag bool `json:"tag,omitempty" mapper:"alias=onTag"`

	// Build image with no cache
	NoCache bool `json:"noCache,omitempty"`

	// Timeout describes how long the build can run
	Timeout *metav1.Duration `json:"timeout,omitempty" mapper:"duration"`

	// Watch describe if a git watcher should be created to watch git branch changes and apply change
	Watch bool `json:"watch,omitempty"`
}
