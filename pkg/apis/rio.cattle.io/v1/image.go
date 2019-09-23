package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ImageBuild struct {
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
	DockerFile string `json:"dockerFile,omitempty"`

	// Specify build context. Defaults to "."
	BuildContext string `json:"buildContext,omitempty"`

	// Specify build args
	BuildArgs []string `json:"buildArgs,omitempty"`

	// Specify the build template. Defaults to `buildkit`.
	Template string `json:"template,omitempty"`

	// Specify the github secret name. Used to create Github webhook, the secret key has to be `accessToken`
	GithubSecretName string `json:"githubSecretName,omitempty"`

	// Specify secret name for checking our git resources
	GitSecretName string `json:"gitSecretName,omitempty"`

	// Specify custom registry to push the image instead of built-in one
	PushRegistry string `json:"pushRegistry,omitempty"`

	// Specify secret for pushing to custom registry
	PushRegistrySecretName string `json:"pushRegistrySecretName,omitempty"`

	// Specify image name instead of the one generated from service name, format: $registry/$imageName:$revision
	BuildImageName string `json:"buildImageName,omitempty"`

	// Whether to enable builds for pull requests
	PR bool `json:"pr,omitempty"`

	// Whether to enable builds for pull requests
	Tag bool `json:"tag,omitempty"`

	// Build image with no cache
	NoCache bool `json:"noCache,omitempty"`

	// BuildTimeout describes how long the build can run
	BuildTimeout *metav1.Duration

	// ServiceName The service to update when the build succeeds or fails.  If blank do not update any service.
	ServiceName string `json:"serviceName,omitempty"`
}

type ImageBuildStatus struct {
}
